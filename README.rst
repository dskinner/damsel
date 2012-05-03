This project is a work in progress. Refer to /tests for syntax or read http://dmsl.dasa.cc for a general idea (minus all the python parts)

damsel
======
For the Go language, damsel allows you to write short notation using css selectors, with the minor addition of a % prefixing an html tag.

::

  %html %body
    #content.border Hello, World

Calling damsel.ParseFile(filename string) will automatically parse the template, pass it on to html/template, and return an instance of that. From there you can call the method Execute(data interface{}). For example, given the data [10][10]int

::

  %html %body
    %table {range .}
      %tr {range .}
      \{end}
    \{end}

While {end} is currently required, in the future this should hopefully be lifted. Also note the escape before {end} is relevant for multiline plain text, for example

::

  %p One
    \ Two
    \ Three

  %p One Two Three

Both paragraphs will be rendered the same.

For more exmaples, refer to the templates under /tests or read http://dmsl.dasa.cc for a general idea (minus all the python parts)
