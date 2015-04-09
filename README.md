# Damsel [![Build Status](https://drone.io/github.com/dskinner/damsel/status.png)](https://drone.io/github.com/dskinner/damsel/latest) [![GoDoc](https://godoc.org/dasa.cc/damsel?status.svg)](https://godoc.org/dasa.cc/damsel)

Markup language featuring html outlining via css-selectors, extensible via pkg html/template and others.

## Library

This package expects to exist at `$GOPATH/src/dasa.cc/damsel` and can be installed with:

```
go get dasa.cc/damsel
```

## Command Line

A command line utility can be installed with:

```
go get dasa.cc/damsel/cmd/damsel
```

View help with:

```
damsel -h
```

## Documentation

http://godoc.org/dasa.cc/damsel

Package damsel provides html outlining via css-selectors and common template functionality.

### Tags

Tags are specified with %tag, #id, .class where #id and .class become a div
if no %tag is specified. Multiple classes can be specified but only one #id
tag should be specified for an element. It's also important that an #id tag
is unique in the document as damsel facilitates overriding content of a
document via an #id tag.

	%html %body
	  #content.border Hello, World
	  %div.one.two.three

### Attributes

Attributes can be inlined, line-breaked, or a combination of such. Provide
quotes around the attribute value to escape brackets.

	%html %body

	  %div[a=1][b=2]

	  #foo
	    [a="1[1]"]
	    [b="2[2]"]

	  #bar[a=1][b=2]
	      [c=3][d=4]

	  %span[a][b] Attributes do not require values

### Text and Whitespace

Whitespace can be manipulated as described below, but it's worth pointing out that
large amounts of content are simply not suitable for such document types (Damsel, Haml, etc).

	%p One
	  \ Two
	  \ Three

	%p    One Two Three

Both paragraphs will be rendered the same. Use of a backslash controls
whitespace, including inlined text.

	%p \ One Two Three

This would insert a space before One.

Whitespace can be preserved using `

	%p `this is some
	text and all whitespace
	    is preserved as-is`

### HTML Comments

Supports commenting out blocks of code via html comments with optional
text and browser specfic IFs. This also includes DOCTYPE declarations.

	!DOCTYPE html
	%html %body
	  ! %ul
	    %li 1
	    %li 2

The use of ! as a block element, causing %ul to become inlined, will cause the
entire block to become commented out. The ! could also be placed after the
body tag in this case. You can also nest items under a comment.

	%html %body
	  ! here's a comment though it's not required
	    %h1 Hello World

	  %div
	    ! this comment doesn't enclose any tags
	    %span Hello World

	  ![if IE] %p Internet Explorer

### Actions

There is basic support for actions. An action is just another way of calling a function
while also preserving indention of the inner content lines, making this suitable for
parsing other indention based markup. Once an action has been processed, the lexer will parse
the result as though it was part of the original document.

In time, this package will facilitate custom functions. Currently
included actions are js, css, include, and extends.

	%html %head
	  :css /css/
	    main.css
	    extra.css

This would generate the following document.

	%html %head
	  %link[type=text/css][rel=stylesheet][href=/css/main.css]
	  %link[type=text/css][rel=stylesheet][href=/css/extra.css]

### Reusable Templates

Damsel allows any element with an id specified to be overridden. Also required
is at least one root node that will serve as the main document output.
Additional root nodes are checked against the first for overridable content.

	%html %body
	  #content
	    %p Hello

	#content OVERRIDE

The would produce the following output.

	<html><body>
	  <div id="content">OVERRIDE</div>
	</body></html>

This functionality is also facilitated with the action extends.

	:extends overlay.dmsl

	#content OVERRIDE

If you wish to append child nodes instead of overriding the original content, specify a super attribute.

	:extends overlay.dmsl

	#content[super]
	  %p A second paragraph

The action include uses the whitespace preceding its declaration to insert
content from a separate document into the current. For example, given the
following document:

	%ul
	  %li One
	  %li Two

included in the following document:

	%html %body
	  #content
	    %h1 My Numbers
	    :include numbers.dmsl

would produce the following document:

	%html %body
	  #content
	    %h1 My Numbers
	    %ul
	      %li One
	      %li Two

### Other Template Integration

This package should be ok for use with most text templating options. Helpers
that provide integration with html/template take the following steps.

	- Call dmsl/parse.ActionParse(b []byte)
	- Pass the result to package html/template and execute
	- Call dmsl/parse.DocParse(b []byte)
	- Display result

Calling dmsl.NewHtmlTemplate will return an instance that parses actions and passes the result
to html/template.Template. Call Execute(interface{}) to produce the final result. For example, given [10][10]int:

	%html %body
	  %table {range .}
	    %tr {range .}
	      %td {.}
	{end}{end}

At one-point the {end} was optional with deeper integration of html/template, but in practice this created
confusion and errors except for the most trivial of examples (above).

Here's another example.

	%html %body
	  %table
	  {range .}

	    {if .}

	    %tr
	    {range .}
	      %td {.}
	    {end}

	    {else}
	    %tr %td None
	    {end}

	    %p some trailing text
	  {end}
