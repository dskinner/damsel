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

# Usage

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
	// TemplateDir is optional, and eventually will *not* be specified globally
	dmsl.TemplateDir = "path/to/dir"
	t := dmsl.ParseFile("path/to/file/in/TemplateDir")
	result, err := t.Execute(nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

```

## Examples

### Tags & Text

```
%html %body
	#content.border Hello, World
	%div.one.two.three
```

Tags are specified with %tag, #id, .class where #id and .class become a div if no %tag is specified. Multiple classes can be specified but only one #id tag should be specified for an element.
It's also important that an #id tag is unique in the document as damsel facilitates overriding content of a document via an #id tag.

### Attributes

```
%html %body

	%div[a=1][b=2]

	#foo
		[a="1[1]"]
		[b="2[2]"]

	#bar[a=1][b=2]
		[c=3][d=4]
```

Attributes can be inlined, line-breaked, or a combination of such. Provide quotes around the attribute value to escape [].

### Text & Whitespace

```
	%p One
		\ Two
		\ Three

	%p    One Two Three
```

Both paragraphs will be rendered the same. Use a backslash to control whitespace at the beginning of a line, including inlined text.

```
	%p \ One Two Three
```

This would insert a space before One.

Also keep in mind that trailing whitespace is currently preserved but subject to change!

### Integration with html/template

Calling dmsl.ParseFile or dmsl.Parse will return an instance of dmsl.Template that has already parsed the dmsl and loaded the result in an html/template.Template.
Call the Execute(interface{}) method from there, given [10][10]int for example:

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




### Damsel Comments

```
%html %body
	// this is a damsel comment, look familiar?
	// comments can only exist on new lines
		// and indention is irrelevant for comments
	/ technically this is a comment but may not be supported
	/ in the future so don't do this
	%p one two three
		\ four five six
		// they can be placed anywhere as long as its a new line
		\ seven eight nine
		\// this isn't a comment and will get appended as text
```

### HTML Comments

There's basic support for commenting out blocks of code via html comments, and also comment text. Support for browser specific IFs is on my TODO

```
%html %body
	! %ul
		%li 1
		%li 2
```
The use of ! as a block element, causing %ul to become inlined will cause the entire block to become commented out. The ! could also be placed after the body tag in this case. You can also nest items under a comment.

```
%html %body
	! here's a comment though it's not required
		%h1 Hello World
	%div
		! this comment doesn't enclose any tags
		%span Hello World
```

### More Examples
For more exmaples, refer to the templates under /tests or read http://dmsl.dasa.cc for a general idea (minus all the python parts)
