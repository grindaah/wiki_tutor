package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "html/template"
    "path"
    "os"
    "regexp"
)

// TODO I am not using it anymore, just kept as an example for multiline const 
const (
    edit_tmpl = `<h1> Editing %s</h1>
    <form action="/save/%s" method="POST">
        <textarea name="body"> %s </textarea> <br>
        <input type="submit" value="Save">
    </form>`
    dataPath = "data"
    tmplPath = "tmpl"    
)

//TODO: some special sympbols also?
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+$)")
//TODO: move templates to tmpl
var templatesWiki = template.Must(template.ParseFiles("tmpl/edit_tmpl.html", "tmpl/view_tmpl.html"))

func init() {
    os.Mkdir(dataPath, 0760)
}

type Page struct {
    Title string
    Body []byte
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    //tmpl = path.Join(tmplPath, tmpl)
    err := templatesWiki.ExecuteTemplate(w, tmpl+".html", p)
    if (err != nil) {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    fmt.Println("viewHandler")
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view_tmpl", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    fmt.Println("editHandler: ", title)
    p, err := loadPage(title)
    
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit_tmpl", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    fmt.Println (body)
    p := &Page{Title: title, Body: []byte (body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w,r, "/view/"+ title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func (p* Page) save() error {
    // TODO why Page shoul be aware about path we are placing it?
    filename := path.Join(dataPath, p.Title)
    filename = filename + ".txt"
    fmt.Printf("saving %s to file: %s\n", p.Title, filename)
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    fmt.Println("Loading page: ", title)
    filename := path.Join(dataPath, title)
    filename = filename + ".txt"
    body, err := ioutil.ReadFile(filename)

    if err != nil {
        return nil, err
    }
    fmt.Println("body:", body)
    return &Page{Title: title, Body: body}, nil
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    log.Fatal(http.ListenAndServe(":8080",nil))
}