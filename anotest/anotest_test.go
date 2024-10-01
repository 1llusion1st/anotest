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

		at.Chapter("image", "image demo", func(t *testing.T) {
			at.PutD2Svg("alice -> bob -> molly")
		})

		at.Chapter("chapter1", "simple title for chapter1", func(t *testing.T) {

			require.Equal(t, 1, 1)

			at.Comment("here are some subchapters ...")

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

			at.StartCode("some code sample")

			fmt.Printf("hello world!")

			for i := 0; i < 100; i++ {
				go func() {
					fmt.Printf("0")
				}()
			}

			at.StopCode()
		})
	})
}
