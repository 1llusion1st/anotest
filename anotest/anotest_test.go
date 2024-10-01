package anotest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNative(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		t.Run("c1", func(t *testing.T) {
			t.Run("s1", func(t *testing.T) {})
			t.Run("s2", func(t *testing.T) {})
		})
		t.Run("c2", func(t *testing.T) {
			t.Run("s1", func(t *testing.T) {})
		})
	})
}

func TestAnnotestTest(t *testing.T) {
	at, err := NewAnotateTest(t, "~/notes/anotest.demo.md", WithDuration())
	require.NoError(t, err)

	at.Story("showcase", func(t *testing.T) {

		at.Chapter("diagram show case", "image demo", func(t *testing.T) {
			at.PutD2Svg(`
				shape: sequence_diagram

				a -> b: hello
				b -> c: should reply ?
				c -> b: I think yes
				b -> a: hi there!
			`)

			at.StartCapture("capture1", "some code sample")

			fmt.Printf("hello world!\n")

			for i := 0; i < 10; i++ {

				go func(i int) {
					fmt.Printf("line %d\n", i)
				}(i)
			}

			at.StopCapture()
		})

		at.Chapter("chapter1", "simple title for chapter1", func(t *testing.T) {

			require.Equal(t, 1, 1)

			at.Comment("here are some subchapters ..." + at.ML("go to listing 1", "capture1"))

			at.Chapter("sub1", "some subchupter 1.1", func(t *testing.T) {
				require.Equal(t, 1, 1)

				at.Comment("some subchapter comment")

				at.Comment(strings.Repeat("some other comment .....", 30))
			})

			at.Chapter("sub2", "some subchapter 1.2", func(t *testing.T) {

				at.Comment(strings.Repeat("some others comment .....", 15))

				require.Equal(t, 1, 1)
			})

			at.Chapter("sub3", "some subchapter 1.3", func(t *testing.T) {
				at.Comment(strings.Repeat("some others comment .....", 15))

				require.NotEqual(t, 1, 2)
			})

			at.Chapter("sub4", "some subchapter 1.4", func(t *testing.T) {

				require.Equal(t, 1, 1)
			})

			at.Comment("and here you can see another comment link" + MakeLink("ch1.sub2", "sub2")).
				Br().
				PutD2Svg("a -> b")

		})

		at.Chapter("another", "some chapter 2", func(t *testing.T) {
			require.Equal(t, 2, 2)

		})
	})
}
