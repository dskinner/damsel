This project is a work in progress and highly experimental with practically no real world use, yet! I have one project I'll be using this on which should help fix most parser errors and I love bug reports. Refer to /tests for syntax or read http://dmsl.dasa.cc for a general idea (minus all the python parts)

damsel
======
For the Go language, damsel allows you to write short notation using css selectors, with the minor addition of a % prefixing an html tag.

Syntax
======
The syntax for damsel is carried over from my python project dmsl. In that case, % is used at the beginning of tags since python could be embedded without any special syntax.

In the case of this Go implementation, I'm giving special consideration to the choices I made for dmsl to see if they are still valid here. Feel free to share your thoughts on the subject by creating a new issue.

Examples
========

::

  %html %body
    #content.border Hello, World

Calling damsel.ParseFile(filename string) will automatically parse the template, pass it on to html/template, and return an instance of that. From there you can call the method Execute(data interface{}). For example, given the data [10][10]int

::

  %html %body
    %table {range .}
      %tr {range .}
        %td {.}

The {end} here is optional (omitted in the example above), but there are times you may wish to use {end} to properly control flow

::

  %html %body
    %table {range .}
      {if .}
      %tr {range .}
        %td {.}
      {else}
      %tr %td None
      {end}

Multiline Plain Text
--------------------

::

  %p One
    \ Two
    \ Three

  %p One Two Three

Both paragraphs will be rendered the same.

For more exmaples, refer to the templates under /tests or read http://dmsl.dasa.cc for a general idea (minus all the python parts)
