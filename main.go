package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// LanguageRegistry maps file extensions to tree-sitter languages.
var LanguageRegistry = map[string]*sitter.Language{
	".go":   golang.GetLanguage(),
	".py":   python.GetLanguage(),
	".js":   javascript.GetLanguage(),
	".ts":   typescript.GetLanguage(),
	".tsx":  typescript.GetLanguage(),
	".rs":   rust.GetLanguage(),
}

// SearchResult represents the structured output for the AI agent.
type SearchResult struct {
	Symbol      string `json:"symbol" xml:"-"`
	NodeType    string `json:"node_type" xml:"node_type,attr"`
	File        string `json:"file" xml:"file,attr"`
	StartLine   int    `json:"start_line" xml:"start_line,attr"`
	EndLine     int    `json:"end_line" xml:"end_line,attr"`
	ParentBlock string `json:"parent_block" xml:"parent_block"`
	Code        string `json:"code" xml:"code"`
}

// XMLResults is the root element for XML output.
type XMLResults struct {
	XMLName      xml.Name       `xml:"search_results"`
	Symbol       string         `xml:"symbol,attr"`
	TotalMatches int            `xml:"total_matches,attr"`
	Matches      []SearchResult `xml:"match"`
}

var (
	dirFlag     = flag.String("dir", ".", "target search directory")
	symbolFlag  = flag.String("symbol", "", "target symbol to search")
	typeFlag    = flag.String("type", "", "specific AST node type (optional)")
	formatFlag  = flag.String("format", "json", "output format (json|xml)")
	queryFlag   = flag.String("query", "", "custom tree-sitter S-expression query")
	maxTokens   = flag.Int("max-tokens", 500, "approximate max tokens for code fragment (4 chars/token)")
	excludeFlag = flag.String("exclude", "", "comma-separated glob patterns to exclude")
)

var binaryExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".exe":  true,
	".so":   true,
	".zip":  true,
	".pdf":  true,
	".bin":  true,
	".pyc":  true,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "agtgrep: AI-optimized code search engine\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [dir] [symbol]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Strict precedence: Flags first, then positional fallbacks
	searchDir := *dirFlag
	searchSymbol := *symbolFlag

	args := flag.Args()

	// Only evaluate positional arguments if flags didn't supply them
	if searchSymbol == "" {
		if len(args) == 1 {
			searchSymbol = args[0]
		} else if len(args) >= 2 {
			if searchDir == "" || searchDir == "." {
				searchDir = args[0]
			}
			searchSymbol = args[1]
		}
	}

	// Final fallback for searchDir
	if searchDir == "" {
		searchDir = "."
	}

	if searchSymbol == "" {
		fmt.Fprintln(os.Stderr, "Error: search symbol is required")
		flag.Usage()
		os.Exit(1)
	}

	excludePatterns := []string{}
	if *excludeFlag != "" {
		excludePatterns = strings.Split(*excludeFlag, ",")
	}

	results := []SearchResult{}

	// Load .gitignore if it exists in searchDir
	var gitIgnore *gitignore.GitIgnore
	gitignorePath := filepath.Join(searchDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		gitIgnore, _ = gitignore.CompileIgnoreFile(gitignorePath)
	}

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if path == searchDir {
				return err
			}
			return nil
		}

		relPath, _ := filepath.Rel(searchDir, path)

		// 1. Skip .git, node_modules, vendor and hidden dirs
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || (len(name) > 1 && strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
		}

		// 2. Skip binary files by extension
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if binaryExtensions[ext] {
				return nil
			}
		}

		// 3. Check .gitignore
		if gitIgnore != nil && gitIgnore.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 4. Check custom exclude patterns
		for _, pattern := range excludePatterns {
			matched, _ := filepath.Match(strings.TrimSpace(pattern), info.Name())
			if !matched {
				// Also check if pattern matches the relative path
				matched, _ = filepath.Match(strings.TrimSpace(pattern), relPath)
			}
			if matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		lang, ok := LanguageRegistry[ext]
		if !ok {
			return nil
		}

		fileResults, err := processFile(path, lang, searchSymbol, *queryFlag)
		if err != nil {
			return nil
		}
		results = append(results, fileResults...)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if strings.ToLower(*formatFlag) == "xml" {
		xmlResults := XMLResults{
			Symbol:       searchSymbol,
			TotalMatches: len(results),
			Matches:      results,
		}
		os.Stdout.WriteString(xml.Header)
		encoder := xml.NewEncoder(os.Stdout)
		encoder.Indent("", "  ")
		if err := encoder.Encode(xmlResults); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding XML: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
	} else {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	}
}

func processFile(path string, lang *sitter.Language, symbol, queryString string) ([]SearchResult, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree := parser.Parse(nil, content)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse")
	}
	root := tree.RootNode()
	if root == nil {
		return nil, fmt.Errorf("nil root")
	}

	results := []SearchResult{}

	if queryString != "" {
		query, err := sitter.NewQuery([]byte(queryString), lang)
		if err != nil {
			return nil, fmt.Errorf("invalid query: %v", err)
		}
		cursor := sitter.NewQueryCursor()
		cursor.Exec(query, root)

		for {
			match, ok := cursor.NextMatch()
			if !ok {
				break
			}
			for _, capture := range match.Captures {
				node := capture.Node
				nodeType := node.Type()
				res := SearchResult{
					Symbol:    symbol,
					NodeType:  nodeType,
					File:      path,
					StartLine: int(node.StartPoint().Row) + 1,
					EndLine:   int(node.EndPoint().Row) + 1,
					Code:      getTrimmedCode(node, content, *maxTokens),
				}
				res.ParentBlock = getParentSignature(node, content)
				results = append(results, res)
			}
		}
	} else {
		searchInNode(root, content, path, symbol, &results)
	}

	return results, nil
}

func searchInNode(node *sitter.Node, content []byte, path, symbol string, results *[]SearchResult) {
	nodeType := node.Type()

	isIdentifier := strings.Contains(nodeType, "identifier") ||
		nodeType == "field_identifier" ||
		nodeType == "type_identifier" ||
		nodeType == "function_declarator"

	if isIdentifier {
		nodeText := string(content[node.StartByte():node.EndByte()])
		if nodeText == symbol {
			if *typeFlag == "" || strings.Contains(strings.ToLower(nodeType), strings.ToLower(*typeFlag)) {
				res := SearchResult{
					Symbol:    symbol,
					NodeType:  nodeType,
					File:      path,
					StartLine: int(node.StartPoint().Row) + 1,
					EndLine:   int(node.EndPoint().Row) + 1,
					Code:      getTrimmedCode(node, content, *maxTokens),
				}
				res.ParentBlock = getParentSignature(node, content)
				*results = append(*results, res)
			}
		}
	}

	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		searchInNode(node.Child(i), content, path, symbol, results)
	}
}

func getParentSignature(node *sitter.Node, content []byte) string {
	curr := node.Parent()
	for curr != nil {
		t := curr.Type()
		if strings.Contains(t, "declaration") || strings.Contains(t, "type") || strings.Contains(t, "literal") || strings.Contains(t, "class") || strings.Contains(t, "method") {
			fullText := string(content[curr.StartByte():curr.EndByte()])
			lines := strings.Split(fullText, "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
		curr = curr.Parent()
	}
	return "global"
}

func getTrimmedCode(node *sitter.Node, content []byte, tokens int) string {
	target := node
	if node.Parent() != nil {
		p := node.Parent()
		pt := p.Type()
		if !strings.Contains(pt, "source_file") && !strings.Contains(pt, "block") && p.EndByte()-p.StartByte() < uint32(tokens*4) {
			target = p
		}
	}

	code := string(content[target.StartByte():target.EndByte()])
	code = strings.TrimSpace(code)

	charLimit := tokens * 4
	if len(code) > charLimit {
		code = code[:charLimit] + "... [truncated]"
	}
	return code
}
