package main

import (
	"bufio"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"os"
	"os/exec"
	"strings"
)

type selpg_args struct {
	start_page int
	end_page int
	in_filename string
	page_len int
	page_type bool /* True for lines-delimited, False for form-feed-delimited, Default is true */
	print_dest string
}

func validate_args(sa selpg_args, rest int) { // 检验输入参数是否合法，rest为剩余的参数数目
	if rest > 1 {
		fmt.Fprintf(os.Stderr, "./selpg: too much arguments\n")
		usage()
		os.Exit(1)
	}
	if sa.start_page <= 0 || sa.end_page <= 0 || sa.end_page < sa.start_page {
		fmt.Fprintf(os.Stderr, "./selpg: Invalid start, end page or line number")
		usage()
		os.Exit(1)
	}
	if sa.page_type == false && sa.page_len != -1 {
		fmt.Fprintf(os.Stderr, "./selpg: Conflict flags: -f and -l")
		usage()
		os.Exit(1)
	}
}

func process_input(sa selpg_args) {
	// initial
	fin := os.Stdin
	fout := os.Stdout
	line_ctr := 0 /* line counter */
	page_ctr := 1 /* page counter */
	var inpipe io.WriteCloser
	var err error

	if sa.in_filename != "" {
		fin, err = os.Open(sa.in_filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "./selpg: could not open input file \"%s\"\n", sa.in_filename)
			usage()
			os.Exit(1)
		}
		defer fin.Close()        // 函数返回前执行fin.Close()
	}

	if sa.print_dest != "" {
		cmd := exec.Command("lp", "-d", sa.print_dest)
		inpipe, err = cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open pipe to \"%s\"\n", sa.print_dest)
			usage()
			os.Exit(1)
		}
		defer inpipe.Close()
		cmd.Stdout= fout
		cmd.Start()
	}

	if sa.page_type == true {
		line := bufio.NewScanner(fin)

		for line.Scan() {
			if page_ctr >= sa.start_page && page_ctr <= sa.end_page  {
				fout.Write([]byte(line.Text() + "\n"))
				if sa.print_dest != "" {
					inpipe.Write([]byte(line.Text() + "\n"))
				}
			}
			line_ctr++
			if line_ctr == sa.page_len {
				page_ctr++
				line_ctr = 0
			}
		}
	} else {
		reader := bufio.NewReader(fin)
		for {
			pageContent, err := reader.ReadString('\f')
			if err != nil || err == io.EOF {
				if err == io.EOF {
					if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
						fmt.Fprintf(fout, "%s", pageContent)
					}
				}
				break
			}

			pageContent = strings.Replace(pageContent, "\f", "", -1)
			if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
				fmt.Fprintf(fout, "%s", pageContent)
			}
			page_ctr++
		}
	}
	if page_ctr < sa.start_page {
		fmt.Fprintf(os.Stderr, "./selpg:  start_page (%d) greater than total pages (%d), less output than expected\n", sa.start_page, page_ctr)
	} else if page_ctr < sa.end_page {
		fmt.Fprintf(os.Stderr, "./selpg:  end_page (%d) greater than total pages (%d), less output than expected\n", sa.end_page, page_ctr)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg --s start_page --e end_page [ --f | --l lines_per_page ] [ --d dest ] [ in_filename ]\n")
}

func main() {
	sa := new(selpg_args)

	// Get args by flag(Pflag)

	flag.IntVar(&sa.start_page, "s", -1, "The start page")
	flag.IntVar(&sa.end_page, "e", -1, "The end page")
	flag.IntVar(&sa.page_len, "l", -1, "The length of the page")
	flag.StringVar(&sa.print_dest, "d", "", "The destination to print")

	f_flag := flag.Bool("f", false, "")
	flag.Parse()

	if *f_flag {
		sa.page_type = false
		sa.page_len = -1
	} else {
		sa.page_type = true  // page_type default True
	}


	sa.in_filename = ""

	if flag.NArg() == 1 {
		sa.in_filename = flag.Arg(0)
	}

	validate_args(*sa, flag.NArg())
	process_input(*sa)
}


