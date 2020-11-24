package main

/*Manga type*/
type Manga struct {
	SourceURL string
	Name      string
	Chapters  []Chapter
}

type Chapter struct {
	Number int
	Title  string
	Pages  []Page
}

type Page struct {
	Number   int
	ImageURL string
}
