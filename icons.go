package main

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

// The icons are Lucide's (ISC license, data/icons/LICENSE), drawn at a
// 1.75 stroke.
//
//go:embed data/icons/*.svg
var iconFiles embed.FS

func icon(name string) *ui.SVG {
	data, err := iconFiles.ReadFile("data/icons/" + name + ".svg")
	if err != nil {
		panic(err)
	}
	return ui.MustParseSVG(data)
}

var (
	iconHome     = icon("house")
	iconRead     = icon("book-open")
	iconShuffle  = icon("shuffle")
	iconStats    = icon("chart-column")
	iconListen   = icon("headphones")
	iconSettings = icon("settings")
	iconPlay     = icon("play")
	iconPause    = icon("pause")
	iconPrev     = icon("skip-back")
	iconNext     = icon("skip-forward")
	iconStop     = icon("square")
	iconBack     = icon("arrow-left")
	iconForward  = icon("arrow-right")
	iconChevron  = icon("chevron-down")
	iconCheck    = icon("check")
	iconDownload = icon("download")
	iconTrash    = icon("trash-2")
	iconClose    = icon("x")
	iconStreak   = icon("flame")
	iconSound    = icon("volume-2")
	iconAdd      = icon("plus")
	iconMinus    = icon("minus")
	iconShare    = icon("share-2")
	iconCopy     = icon("copy")
	iconRepeat   = icon("repeat")
	iconText     = icon("type")
	iconBook     = icon("book")
	iconSparkles = icon("sparkles")
	iconList     = icon("list")
)

// iconButton is a button showing an icon alone, named for assistive
// technology and in a tooltip.
func iconButton(c *ui.Context, svg *ui.SVG, label string) ui.Element {
	return ui.Button(c.Key(label), "").Label(label).Tooltip(label).Children(func() {
		ui.Icon(c, svg)
	})
}

// labeledButton is a button with an icon before its label; primary makes
// it the accent's.
func labeledButton(c *ui.Context, svg *ui.SVG, label string, primary bool) ui.Element {
	var b ui.Element
	if primary {
		b = ui.PrimaryButton(c.Key(label), "")
	} else {
		b = ui.Button(c.Key(label), "")
	}
	return b.Label(label).Children(func() {
		ui.Icon(c, svg)
		ui.Text(c, label).SingleLine()
	})
}
