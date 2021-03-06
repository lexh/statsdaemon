// This tool looks at a log generated by `ngrep -qd any . udp dst port 8125 >> <file>` and
// tells you all lines that are invalid.  If you see invalid lines in statsdaemon (or other statsd implementations)
// then this tool helps in making sure the statsd server itself doesn't corrupt the data.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/vimeo/statsdaemon/udp"
	"io"
	"os"
)

func printInvalid(path string) {
	fd, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The file %s does not exist!\n", path)
		return
	}
	defer fd.Close()

	bufReader := bufio.NewReader(fd)
	line_no := 1
	for line, isPrefix, err := bufReader.ReadLine(); err != io.EOF; line, isPrefix, err = bufReader.ReadLine() {
		if isPrefix {
			fmt.Printf("ERROR: Line %d too long to fit in buffer", line_no)
		} else {
			line_no += 1
			if bytes.HasPrefix(line, []byte("U ")) {
				// udp packet header by ngrep
				continue
			}
			if len(line) == 0 {
				// empty line in ngrep output
				continue
			}
			if len(line) == 1 {
				fmt.Println(string(line), "WTF not sure what the error is")
				continue
			}
			// every packet starts with 2 spaces, but that's ok: parseLine strips space anyway
			// also, ngrep output sometimes seems to contain whitespace at the end, but parseLine does that as well.
			_, err := udp.ParseLine2(line)
			if err != nil {
				fmt.Println(string(bytes.TrimSpace(line)), err)
			}
		}
	}
}

func main() {
	flag.Parse()
	for _, path := range flag.Args() {
		fmt.Printf("File %s\n", path)
		printInvalid(path)
	}
}
