# Overview

`reid` is a collection of tools designed to facilitate full-text searches of
academic papers maintained in an [EndNote] library for specific words or
phrases. The `reid` project currently consists of the following programs:

* `reid-enxml` parses an XML-formatted library file exported from
EndNote. This may be used to display information contained in the XML file, or
to convert the XML into a "reid project" file that the remainder of the `reid`
tools can work with.
* `reid-convert` iterates through all, or a specified subset, records stored in
a "reid project" file and converts their associated PDFs into "minified"
text files, which can then be searched with the `reid-search` program. Upon
successfully converting files, the "reid project" file is updated to reflect
the location of the "minified" text files.
* `reid-search` searches the previously "minified" text files for one or
more terms specified on the command line. Simple terms and phrases can be
specified, in addition to [Regular Expressions]. The search can be performed
over all files contained in the "reid project" file, or limited by year range,
author, or publication. For each matched input term, this program outputs
the number of occurrences observed in corresponding source material. By default,The
information about matches are printed to the terminal. However, format of this
output can be changed to CSV or JSON, and the data can be written to a file.

[EndNote]: http://endnote.com/
[Regular Expressions]: https://en.wikipedia.org/wiki/Regular_expression#Basic_concepts

# Requirements

The following are required to build and use `reid`

* Linux - Tessaract reportedly misbehaves on OSX, which has not been tested.
* [Go] 1.7 or later. (Earlier versions have not been tested)
* [tesseract]: `tesseract-eng`, `libtesseract`, `libtesseract-dev`
* [poppler] utilities: `pdfimages`, `pdftotext`


[Go]: https://golang.org/
[tesseract]: https://github.com/tesseract-ocr/tesseract
[poppler]: https://poppler.freedesktop.org/releases.html


# Workflow and Usage

In general, one uses the `reid` tools in the order listed above. This section
briefly explains how to use the tools. Information about building the tools from
source is shown in a later section.

For additional usage information, such as available flags and arguments, run
the tools with a `--help` command line option.

## Exporting an Endnote Library as an XML file

First, export entries from your EndNote library to an XML. This can be done
via the `File > Export...` option in EndNote. In the window that appears,
be sure to specify:

 * `XML (*.xml)` as the File Type
 * `All Fields` as the output format

Also note the checkbox that allows you to export only the selected items.
Check this only if you are trying to create a "reid project" file from
a subset of your library.


## Viewing an exported XML's contents

While the exported XML library files are human-readable text, they a bit
dense and a bit tough to review manually. The `reid-enxml` program's
"show" command can be used to summarize the contents of the XML file. For
large library exports, it may take a few seconds to load the XML file.

For the following examples, assume that an EndNote library has been exported
to a `mylib.xml` file.

**List all records contained in the XML**

As this is likely to be a large list, you can pipe the output of `reid-enxml`
to the `less` program and then use the Page Up/Down or arrow keys to navigate
the list.

~~~
$ reid-enxml -x mylib.xml show all | less
~~~

**List all publications**

Perhaps you want a sorted list of all the publications within a library?
This can be done using `show publications` piping the output to the `sort`
program.

~~~
$ reid-enxml -x mylib.xml show publications | sort | less
~~~

**List all authors**

It also possible to list all authors contained in the library. However,
the accuracy of this is only as good as the metadata from PDFs imported
into your EndNote library. It is likely that different publication will specify
names differently -- including or omitting a middle initial, or possibly
including a full first name instead of an initial.

Listing and sorting authors' names may be helpful in determining if this is the
case in your library.

~~~
$ reid-enxml -x mylib.xml show authors | sort
~~~

**List all years covered by the library contents**

When surveying the use of terms over a period of time, it is often useful
to first confirm that you've exported the correct date ranges before starting
your searches. Use the `show years` command to quickly verify that years
that the exported library covers.

~~~
$ reid-enxml -x mylib.xml show years
~~~


## Creating a reid project file

Only a subset of information from the EndNote library XML file is required.
Additionally, the `reid` tools need to know some information, such as
where your converted text files are (to be) stored. For that reason, a "reid
project" file and an associated data directory must be provided.

The following command converts a `mylib.xml` to a `myproject.json` file and
creates a `mydata` directory. This directory will be empty for now, but will
later be used to stored the text extracted from PDF files.

~~~
$ reid-enxml -x mylib.xml create myproject.json mydata
~~~

## Converting PDFs to "minified" text files

Before being able to search PDF documents with `reid`, we must first extract
the text from them. When converting to text, `reid-convert` also "minifies"
this text to make it easier to search. In the context of the `reid` project,
this "minification" consists of:

* Reformatting text to a single line and re-joining hyphenated text.
* Removing references, URLs, and punctuation.
* Converting any remaining white space to a single space (" ").
* Converting text to lower-case. (As a result, searches are case-insensitive.)

Currently, all of the above is performed by default. If this interferes
with your ability to search a document, please submit a feature request
on the [Issue Tracker] with respect to tuning the various "minifications"
that are performed.

[Issue Tracker]: https://github.com/jynik/reid/issues

The following command will convert all the PDFs associated with entries
in the previously created project file to "minified" text, storing these
text files in the `mydata` directory.

The `--debug` argument is optional; you can specify this to view some
additional information about what the program is currently doing. The process
generally takes a few seconds (or less) per PDF, so depending upon the
size of your EndNote library, this may be a good time to go make yourself
a cup of your favorite hot beverage.

~~~
$ reid-convert -p myproject.json --debug
~~~

Note that `reid-convert` also supports converting only a specified set
of PDF files. This is is largely for debugging purposes and is not expected to
be terribly useful to "end users." Run `reid-convert --help` for the
available options for this.

## Finally...Searching!

With all that done, we finally search our entire library for various
terms and phrases. Below is a simple and fictional example, in which we search
for all the occurrences of the term, "bootloader":

~~~
$ reid-search -p myproject.json -t bootloader
~~~

By default, the results are "pretty printed" to the console:

~~~
Query: bootloader
  Occurrences: 7
  Year: 2015
  Publication: The Journal Of The Internet Of Things That Shouldn't Be
  Author(s): Goodspeed, T. / Ridley, S. / Grand, J. / Fitz, J.
  Title: Cross-Platform ROP Gadget Polyglots for ARM, MIPS, and PIC32

Query: bootloader
  Occurrences: 13
  Year: 2021
  Publication: Phrack #70
  Authors(s): Laphroaig, M.
  Title: Inserting Backdoors Into Black Box Firmware For Fun and Profit
~~~

But what if some works use the term "boot loader" (with a space) instead of
"bootloader?"  The same search can be performed, but with an additional term:

~~~
$ reid-search -p myproject.json -t bootloader -t 'boot loader'

Query: bootloader
  Occurrences: 7
  Year: 2015
  Publication: The Journal Of The Internet Of Things That Shouldn't Be
  Author(s): Goudaspeed, T. / Lidrey, S. / Grandious, J. / Ritz, J.
  Title: Cross-Platform ROP Gadget Polyglots for ARM, MIPS, and PIC32

Query: boot loader
  Occurrences: 42
  Year: 1996
  Publication: Real-time and Embedded Systems
  Author(s): Smith, J.
  Title: Modern Boot Loader Design

Query: bootloader
  Occurrences: 13
  Year: 2021
  Publication: Phrack #70
  Authors(s): Laphroaig, M.
  Title: Inserting Backdoors Into Black Box Firmware For Fun and Profit
~~~

The same search as the above can be performed using a [Regular Expression] to
define the same desired search, using the `--regexp/-r` option instead of the
`--term/-t` option:

~~~
$ reid-search -p myproject.json -r 'boot ?loader'

Query: regexp{boot ?loader}
  Occurrences: 7
  Year: 2015
  Publication: The Journal Of The Internet Of Things That Shouldn't Be
  Author(s): Goudaspeed, T. / Lidrey, S. / Grandious, J. / Ritz, J.
  Title: Cross-Platform ROP Gadget Polyglots for ARM, MIPS, and PIC32

Query: regexp{boot ?loader}
  Occurrences: 9
  Year: 1996
  Publication:
  Author(s): Goodspeed, T. / Ridley, S. / Grand, J. / Fitz
  Title: Cross-Platform ROP Gadget Polyglots for ARM, MIPS, and PIC32

Query: regexp{boot ?loader}
  Occurrences: 13
  Year: 2021
  Publication: Phrack #70
  Authors(s): Laphroaig, M.
  Title: Inserting Backdoors Into Black Box Firmware For Fun and Profit
~~~

It is important to note that one "Occurrences" count will be listed for
everything matched by the regular expression, not each individual term
or phrase. This can be very handy in cases where you want to group occurrences
of similar terms or phrases. For example, the following will report the total
number of occurrences of either "architecture" or "architectural".

~~~
$ reid-search -p myproject.json -r 'architechtur(e|al)'
~~~

When using regular expressions, be aware that it's up to you to specify
where whitespace or the beginning/end of a document may occur. In the
previous example, results for "microarchitecure" or "architecure-independent"
would be included. If you only want an exact match for "architecture" then
you would need to specify `(^| )architecture( |$)`.  In fact, this is 
exactly what the `-t/--term` option is doing with the provided text.
(Remember, the minified text is converted to lower case and all white space
is converted to a single space ' ' character.)

In all liklihood, you won't want always to search the entire library.
Instead, you might want to search articles from a specific publication,
or over a limited number of years.

The `--from/-F` and `--to/-T` arguments can be used to specify the earliest
and latest years to include in searches.

The `--publication/-P` argument can be used to specific publication to search.
This argument can be specified multiple times to include multiple publications
in the search.

Searches can also be done by author, using the `--author/-a` argument. This too
can be specified multiple times to include multiple authors in the search.

Below is an example of these arguments in action:

~~~
$ reid-search --from 1990 --to 2001 \
              --publication 'Real-time and Embedded Systems' \
              --publication 'Circuit Cellar' \
              --publication 'IEEE Xplore' \
              --author 'Smith, J.' \
              --author 'Turing, A.' \
              --author 'Lovelace, A.' \
              -r 'boot ?loader'
~~~

The default "pretty print" output may be helpful for getting a quick sense of
results, but is not well-suited for aggregating a large set of results. Instead,
the `--format/-f` argument may be used to change the output format.  This may be used
in conjunction with the `--outfile/-o` argument to write results to a file.

The available formats are:

* `pretty`: Simple descriptive format. Not well-suited for automated parsing. (Default)
* `csv`: Comma separated values with quoted strings. Can be imported into tools like Excel.
* `csv-no-hdr`: Same as `csv` but without a header row
* `json`: Javascript Object Notation. This is the best option if you want to
work with the data programatically.

For more information, run `reid-search --help`.


# Build

Binary releases of this software are not yet provided; the `reid` tools
must be built from source. (You can always ask me politely for a build,
of course!)

Running `make` from the top-level directory will result in the `go get` calls
needed to fetch and build dependencies.

Upon completion the `reid` tools will be located in the top-level directory.
Copy or move these into a location within your `${PATH}`.

# License

This software is released under version 3.0 of the GNU General Public License.
The text of this license may be found in the [COPYING] file.

# Support (or, Lack Thereof)

This software is developed and maintained on an as-needed basis, in the
author's spare time. As such, no support for these tools is officially offered.

Please use the [Issue Tracker] only to report defects in the software. General
support or usage questions will be immediately closed.

With that being said, the author considers himself a fairly decent human being.
If you're really struggling to use these tools, feel free to send and email
(found in the git commit log), and he'll probably find time to lend a helping
hand. Hint: Amazon gift cards and beer money are always appreciated. `;)`


# Disclaimers

"EndNote" is a registered trademark owned by Clarivate Analytics. The author of
the `reid` tools is not affiliated with Clarivate Analytics.

To the best of his knowledge, this software has been developed in a manner that
is consistent with the EndNote End User License Agreement; the `reid` tools only
process user-exported XML files, and do not utilize any Clarivate-owned applications,
libraries, or SDKs. No reverse engineering of the EndNote software was performed to
develop these tools; `reid-enxml` simply parses the self-explanatory,
human-readable, user-exported library XML files.

However, the author is neither a lawyer nor an actor that plays one on TV. The
user of this software is responsible for ensuring their usage of the `reid`
software is consistent with the EndNote End User License Agreement.

The `reid` tools were developed using a very small sample of XML files output
by EndNote X7. It may not adequately support XML files produced by
other versions of the software. Furthermore, the sample XML files used to
develop these tools contained only journal articles. As such, changes may to
the `reid` tools may be required if one's exported library contains other types
of published works.

Finally, the `reid` tools were developed on best-effort basis. The author
takes no responsibility for, and makes no guarantees of, the correctness of the
data output by these tools. Users are ultimately responsible for ensuring the
validity and correctness of their data and results.
