This project is a work in progress and highly experimental with practically no real world use, yet! I have one project I'm using this on which should help fix most parser errors (there seem to be quite a few right now) and I love bug reports. Refer to /tests for syntax or read http://dmsl.dasa.cc for a general idea (minus all the python parts)

# damsel

For the Go language, damsel allows you to write short notation using css selectors, with the minor addition of a % prefixing an html tag.

# Syntax

The syntax for damsel is carried over from my python project dmsl. In that case, % is used at the beginning of tags since python could be embedded without any special syntax.

In the case of this Go implementation, I'm giving special consideration to the choices I made for dmsl to see if they are still valid here. Feel free to share your thoughts on the subject by creating a new issue.

# Installation

```
go get github.com/dskinner/damsel
```

# Examples

Here are some basic examples. Will fill this area out over time.

## Command Line

```
damsel -f path/to/file
```

## Project Import

```
package main

import (
  "github.com/dskinner/damsel/dmsl"
  "fmt"
)

func main() {
	t := dmsl.ParseFile("path/to/file")
  fmt.Println(t.Execute(nil))
}

```

## Basics

Here's a basic template that inlines the body tag, and creates an implied div with id="content" and class="border"

```
%html %body
  #content.border Hello, World
```

Calling dmsl.ParseFile(filename string) will automatically parse the template, pass it on to html/template, and return an instance of that. From there you can call the method Execute(data interface{}). For example, given the data [10][10]int

```
%html %body
  %table {range .}
    %tr {range .}
      %td {.}
```

The {end} here is optional (omitted in the example above), but there are times you may wish to use {end} to properly control flow

```
%html %body
  %table {range .}
    {if .}
    %tr {range .}
      %td {.}
    {else}
    %tr %td None
    {end}
    %p some trailing text
```

## Attributes

Note, string literals for attribute values is on my TODO

```
%html %body

  #content[a=1][b=2] You can specify multiple attributes like this

  #footer
    [a=1]
    [b=2] You can also line break the attributes
    \ and start text above or here
```

## Multiline Plain Text

```
  %p One
    \ Two
    \ Three

  %p One Two Three
```

Both paragraphs will be rendered the same.

## HTML Comments

There's basic support for commenting out blocks of code via html comments. Support for browser specific IFs and comment text is on my TODO

```
%html %body
  ! %ul
    %li 1
    %li 2
```
The use of ! as a block element, causing %ul to become inlined will cause the entire block to become commented out. The ! could also be placed after the body tag in this case. You can also nest items under a comment.

```
%html %body
	!
    %h1 Hello World
```

## More Examples
For more exmaples, refer to the templates under /tests or read http://dmsl.dasa.cc for a general idea (minus all the python parts)
