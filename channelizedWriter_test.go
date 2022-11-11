package ntnns

import (
	"bytes"
	"context"
	"sync"
	"testing"
)

func TestChannelizedWriter(t *testing.T) {
	cases := map[string]struct {
		ctx    context.Context
		wg     *sync.WaitGroup
		input  []string
		output string
	}{
		"all provided": {
			ctx: context.Background(),
			wg:  new(sync.WaitGroup),
			input: []string{
				"line 1\n",
				"line 2\n",
				"line 3\n",
			},
			output: `line 1
line 2
line 3
`,
		},
		"no context": {
			wg: new(sync.WaitGroup),
			input: []string{
				"line 1\n",
				"line 2\n",
				"line 3\n",
			},
			output: `line 1
line 2
line 3
`,
		},
		"no waitgroup": {
			ctx: context.Background(),
			input: []string{
				"line 1\n",
				"line 2\n",
				"line 3\n",
			},
			output: `line 1
line 2
line 3
`,
		},
		"no context no waitgroup": {
			input: []string{
				"line 1\n",
				"line 2\n",
				"line 3\n",
			},
			output: `line 1
line 2
line 3
`,
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			b := new(bytes.Buffer)
			if cas.wg != nil {
				cas.wg.Add(1)
			}

			cw := NewChannelizedWriter(cas.ctx, b, cas.wg)

			for i, line := range cas.input {
				n, err := cw.Write([]byte(line))
				if err != nil {
					t.Errorf("write returned error: %v", err)
					return
				}
				if n != len(line) {
					t.Errorf("write with line %d %q returned bytecount %d, expected %d", i, line, n, len(line))
					return
				}
			}

			cw.Close()

			if cas.wg != nil {
				cas.wg.Wait()
				if done := cw.Done(); !done {
					t.Errorf("done does not return true after wait")
				}
			} else {
				for !cw.Done() {
				}
			}

			if b.String() != cas.output {
				t.Errorf("written content does not match expected content")
				t.Errorf("expected: %s", cas.output)
				t.Errorf("recevied: %s", b.String())
			}

			if err := cw.Error(); err != nil && err != ErrChannelizedWriterClosed {
				t.Errorf("cw.Error returned unexpected error: %v", err)
			}
		})
	}
}
