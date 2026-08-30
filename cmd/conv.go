package cmd

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	chhtml "github.com/alecthomas/chroma/v2/formatters/html"

	d2 "github.com/FurqanSoftware/goldmark-d2"
	gm "github.com/yuin/goldmark"
	gmhighlighting "github.com/yuin/goldmark-highlighting/v2"
	gmextension "github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	mermaid "go.abhg.dev/goldmark/mermaid"

	//d2themescatalog "github.com/FurqanSoftware/goldmark-d2"

	"github.com/dimchansky/utfbom"
)

//go:embed templates/*
var embeded embed.FS

type renderOptions struct {
	UsePagedJs      bool
	ContainsMermaid bool
}

type renderData struct {
	Content string
	Options renderOptions
}

var extlng = map[string]string{}
var styles = map[string]bool{}

func init() {
	rootCmd.Flags().BoolP("show-in-browser", "b", false, "show in browser")
	rootCmd.Flags().BoolP("use-pagedjs", "p", false, "use paged.js for better pdf print, refer to https://pagedjs.org/")
	rootCmd.Flags().StringP("style", "s", "vs", "style, refer to https://xyproto.github.io/splash/docs")
	rootCmd.Flags().StringP("language", "l", "", "language, refer to https://github.com/alecthomas/chroma#supported-languages")

	// create key value dictionary

	// see https://github.com/alecthomas/chroma#supported-languages
	extlng = make(map[string]string)
	extlng[".bat"] = string("shell")
	extlng[".cpp"] = string("cpp")
	extlng[".cs"] = string("csharp")
	extlng[".css"] = string("css")
	extlng[".curl"] = string("curl")
	extlng[".dart"] = string("dart")
	extlng[".diff"] = string("diff")
	extlng[".dockerfile"] = string("dockerfile")
	extlng[".ebnf"] = string("ebnf")
	extlng[".fs"] = string("fsharp")
	extlng[".go"] = string("go")
	extlng[".html"] = string("html")
	extlng[".http"] = string("http")
	extlng[".xhtml"] = string("xhtml")
	extlng[".xml"] = string("xml")
	extlng[".js"] = string("js")
	extlng[".json"] = string("json")
	extlng[".md"] = string("markdown")
	extlng[".php"] = string("php")
	extlng[".py"] = string("python")
	extlng[".perl"] = string("perl")
	extlng[".pl"] = string("perl")
	extlng[".ps1"] = string("powershell")
	extlng[".postgres"] = string("postgres")
	extlng[".postgresql"] = string("postgresql")
	extlng[".pgsql"] = string("pgsql")
	extlng[".md"] = string("markdown")
	extlng[".mssql"] = string("mssql")
	extlng[".mysql"] = string("mysql")
	extlng[".rb"] = string("ruby")
	extlng[".rust"] = string("rust")
	extlng[".rs"] = string("rust")
	extlng[".sh"] = string("shell")
	extlng[".sql"] = string("sql")
	extlng[".sqlite"] = string("sqlite")
	extlng[".ts"] = string("typescript")
	extlng[".yaml"] = string("yaml")
	extlng[".yml"] = string("yaml")

	// see https://xyproto.github.io/splash/docs/
	styles = make(map[string]bool)
	styles["abap"] = true
	styles["algol"] = true
	styles["algol_nu"] = true
	styles["arduino"] = true
	styles["autumn"] = true
	styles["average"] = true
	styles["base16-snazzy"] = true
	styles["borland"] = true
	styles["bw"] = true
	styles["catppuccin-frappe"] = true
	styles["catppuccin-latte"] = true
	styles["catppuccin-macchiato"] = true
	styles["catppuccin-mocha"] = true
	styles["colorful"] = true
	styles["doom-one"] = true
	styles["doom-one2"] = true
	styles["dracula"] = true
	styles["emacs"] = true
	styles["friendly"] = true
	styles["fruity"] = true
	styles["github-dark"] = true
	styles["github"] = true
	styles["gruvbox-light"] = true
	styles["gruvbox"] = true
	styles["hr_high_contrast"] = true
	styles["hrdark"] = true
	styles["igor"] = true
	styles["lovelace"] = true
	styles["manni"] = true
	styles["modus-operandi"] = true
	styles["modus-vivendi"] = true
	styles["monokai"] = true
	styles["monokailight"] = true
	styles["murphy"] = true
	styles["native"] = true
	styles["nord"] = true
	styles["onedark"] = true
	styles["onesenterprise"] = true
	styles["paraiso-dark"] = true
	styles["paraiso-light"] = true
	styles["pastie"] = true
	styles["perldoc"] = true
	styles["pygments"] = true
	styles["rainbow_dash"] = true
	styles["rose-pine-dawn"] = true
	styles["rose-pine-moon"] = true
	styles["rose-pine"] = true
	styles["rrt"] = true
	styles["solarized-dark"] = true
	styles["solarized-dark256"] = true
	styles["solarized-light"] = true
	styles["swapoff"] = true
	styles["tango"] = true
	styles["tokyonight-day"] = true
	styles["tokyonight-moon"] = true
	styles["tokyonight-night"] = true
	styles["tokyonight-storm"] = true
	styles["trac"] = true
	styles["vim"] = true
	styles["vs"] = true
	styles["vulcan"] = true
	styles["witchhazel"] = true
	styles["xcode-dark"] = true
	styles["xcode"] = true
}

func convCmdFunc(cmd *cobra.Command, args []string) {

	var content []byte
	var ext string
	var fileName string
	var documentDir string
	var err error

	//check if args is empty
	if len(args) > 0 {
		// get file name
		fileName = args[0]
	}

	// get string option
	style, _ := cmd.Flags().GetString("style")
	language, _ := cmd.Flags().GetString("language")
	usePagedJS, _ := cmd.Flags().GetBool("use-pagedjs")
	showInBrowser, _ := cmd.Flags().GetBool("show-in-browser")

	// check style
	if !styles[style] {
		style = "vs"
	}

	// check file exists
	if len(fileName) == 0 {

		// try read stdin
		content, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %s\n", err)
			os.Exit(1)
		}

	} else {

		// get file name extension
		ext = filepath.Ext(fileName)
		documentDir = filepath.Dir(fileName)

		// read content
		content, err = readFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file %s, %s\n", fileName, err)
			os.Exit(1)
		}

		// get language
		if language == "" {
			language = extlng[strings.ToLower(ext)]
		}

	}

	// load templates
	txttmpl := template.Must(template.ParseFS(embeded, "templates/head.html", "templates/footer.html", "templates/script.html", "templates/page-md.html", "templates/page-txt.html"))

	// if pure code or markdown file was given
	if language != "" {

		var containsMermaid bool
		var htmlContent []byte

		// when any language, but not markdown
		if language != "markdown" {

			// wrap content as markdown code block
			content = toMarkdown(language, content)

		} else {

			// check if content contains ":::mermaid" or "```mermaid" with regex
			// when ":::mermaid" then replace to "```mermaid"
			prepareMermaid(&content, &containsMermaid)

		}

		// convert markdown to html
		htmlContent = toHtml(content, style)

		// include images ?
		if language == "markdown" {
			htmlContent = replaceImages(documentDir, htmlContent)
		}

		// check conditions if mermaid is used and use pagedjs
		if containsMermaid && usePagedJS {
			// when mermaid is used and pagedjs is used, then prefer mermaid
			usePagedJS = false
		}

		// prepare renderData
		data := renderData{
			Content: string(htmlContent),
			Options: renderOptions{
				ContainsMermaid: containsMermaid,
				UsePagedJs:      usePagedJS,
			},
		}

		// show in browser ?
		if showInBrowser {

			tmpfile, _ := os.CreateTemp("", "m2h-*.htm")
			err = txttmpl.ExecuteTemplate(tmpfile, "page-md.html", data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error executing template: %s\n", err)
				os.Exit(1)
			}

			_ = tmpfile.Close()
			defer func() { _ = os.Remove(tmpfile.Name()) }()

			openBrowser(tmpfile)

		} else {

			err = txttmpl.ExecuteTemplate(os.Stdout, "page-md.html", data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error executing template: %s\n", err)
				os.Exit(1)
			}

		}

	} else {

		// handle raw text
		unescaped := string(content)

		escaped := html.EscapeString(unescaped)

		// replace \t\r\n and two spaces with <br> or &npsp;
		replace_n := -1 //-1 all
		escaped = strings.Replace(escaped, "\r\n", "<br>", replace_n)
		escaped = strings.Replace(escaped, "\n", "<br>", replace_n)
		escaped = strings.Replace(escaped, "\r", "<br>", replace_n)
		escaped = strings.Replace(escaped, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;", replace_n)
		escaped = strings.Replace(escaped, "  ", "&nbsp;&nbsp;", replace_n)

		// prepare renderData
		data := renderData{
			Content: escaped,
			Options: renderOptions{
				ContainsMermaid: false,
				UsePagedJs:      usePagedJS,
			},
		}

		// print simple text embedded in html to stdout
		err = txttmpl.ExecuteTemplate(os.Stdout, "page-txt.html", data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error rendering template: %s\n", err)
			os.Exit(1)
		}

	}
}

func prepareMermaid(content *[]byte, containsMermaid *bool) {
	// check if content contains ":::mermaid" or "```mermaid" with regex
	// when ":::mermaid" then replace to "```mermaid"
	reMatchMermaid := regexp.MustCompile("(:::|```)\\s*mermaid")

	if reMatchMermaid.Match(*content) {

		*containsMermaid = true
		*content = bytes.ReplaceAll(*content, []byte(":::"), []byte("```"))
	}
}

func readFile(fileName string) ([]byte, error) {

	// open file
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// read file, skip bom char if exists
	return io.ReadAll(utfbom.SkipOnly(file))
}

func readStdin() ([]byte, error) {

	// read stdin
	return io.ReadAll(os.Stdin)
}

func toMarkdown(lng string, content []byte) []byte {

	headlng := fmt.Sprintf("```%s\r\n", lng)
	footlng := "\r\n```\r\n"

	markdown := make([]byte, 0)
	markdown = append(markdown, []byte(headlng)...)
	markdown = append(markdown, content...)
	markdown = append(markdown, []byte(footlng)...)

	return markdown
}

func toHtml(content []byte, style string) []byte {

	md := gm.New(
		gm.WithExtensions(
			&mermaid.Extender{
				RenderMode: mermaid.RenderModeClient,
				NoScript:   true,
			},
			&d2.Extender{
				// Defaults when omitted
				// Layout:  d2.Layout,
				// ThemeID: d2.CoolClassics.ID,
			},
			gmextension.GFM,
			gmhighlighting.NewHighlighting(
				gmhighlighting.WithStyle(style),
				gmhighlighting.WithFormatOptions(
					chhtml.WithLineNumbers(true),
				),
			),
		),
		gm.WithParserOptions(
			gmparser.WithAutoHeadingID(),
		),
		gm.WithRendererOptions(
			gmhtml.WithHardWraps(),
			gmhtml.WithXHTML(),
			gmhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func replaceImages(documentDir string, html []byte) []byte {

	content := string(html)

	// It is not possible to use negative lookbehind in golang `<img src="(?!data:image/)([^"]+)"`
	// therefore use two regexps
	// Part 1: Match all strings that do not contain ",
	regexImgTag := regexp.MustCompile(`(?i)<img src="([^"]+)"`)

	// Part 2: Filter out strings that start with data:image/
	regexDataImage := regexp.MustCompile(`(?i)^data:image/`)

	// Part 1: find all `<img src="...">`
	matches := regexImgTag.FindAllStringSubmatch(string(content), -1)

	//iterate through all regexp matches and replace image with base64 encoded data
	for _, match := range matches {

		src := match[1]

		// Part 2: exclude all data:image/ strings
		if !regexDataImage.MatchString(src) {
			fullSrc := filepath.Join(documentDir, src)
			imgExt := strings.ToLower(filepath.Ext(src)[1:]) // remove dot

			switch imgExt {
			// nothing to do ...
			//case "jpg", "jpeg":
			//case "png", "gif", "bmp", "tiff":

			// convert svg to svg+xml
			case "svg":
				imgExt = "svg+xml"
			}

			b64, err := getBase64(fullSrc)
			if err == nil {
				old := fmt.Sprintf(`src="%s"`, src)
				new := fmt.Sprintf(`src="data:image/%s;base64,%s"`, imgExt, b64)

				content = strings.ReplaceAll(content, old, new)
			}
		}

	}

	return []byte(content)
}

func getBase64(fileName string) (string, error) {

	// open file
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	// read file, skip bom char if exists
	image, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// encode to base64
	encoded := base64.StdEncoding.EncodeToString(image)

	return encoded, nil
}

func openBrowser(tmpfile *os.File) {

	// open in browser
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", tmpfile.Name()).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", tmpfile.Name()).Start()
	case "darwin":
		_ = exec.Command("open", tmpfile.Name()).Start()
	default:
		fmt.Println("unsupported platform")
	}

	time.Sleep(4 * time.Second)

}
