package clicolor_test

import cli "github.com/ivpusic/go-clicolor/clicolor"
import "testing"

type WriterFunc func([]byte) (int, error)

var ch chan string

func (f WriterFunc) Write(p []byte) (n int, err error) {
	return f(p)
}

func ExamplePrint_in() {
	cli.Print("this is green text").In("green")
	// prints: this is green text
}

func ExamplePrint_inFormat() {
	cli.Print("{red}Some text in red. {white}Some text in white. {default}Some text in default color.").InFormat()
	// prints: Some text in red. Some text in white. Some text in default color.
}

func Write(p []byte) (n int, err error) {
	str := string(p[:])
	ch <- str
	return 0, nil
}

func TestPrintIn(t *testing.T) {
	ch = make(chan string, 1)
	cli.Out = WriterFunc(Write)
	cli.Print("this is green text").In("green")
	str := <-ch

	expected := "\033[32mthis is green text\033[0m\n"
	if str != expected {
		t.Error("Expected", expected)
	}

	close(ch)
}

func TestPrintInFormat(t *testing.T) {
	ch = make(chan string, 1)
	cli.Out = WriterFunc(Write)
	cli.Print("{red}Some text in red. {white}Some text in white. {default}Some text in default color.").InFormat()
	str := <-ch

	expected := "\033[31mSome text in red. \033[0m\033[37mSome text in white. \033[0m\033[39mSome text in default color.\033[0m\n"
	if str != expected {
		t.Error("Expected", expected, "Got", str)
	}

	close(ch)
}
