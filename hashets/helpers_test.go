package hashets

import "os"

var (
	testdataIn     = os.DirFS("../testdata/in")
	testdataExpect = os.DirFS("../testdata/expect")

	expectMap = Map{
		"folder/maja.webp":     "folder/maja_0dc4c09273ee73a1fdcdcb1b39512ec2322e8115c6738b53e1435838bfb4f9ca.webp",
		"bee movie.txt":        "bee movie_d3bb03cafa9d9f1678173a9e65d9d7c6eca60a3e53a551481e5d9dbe8970d13c.txt",
		"cheesy_fur.ext1.ext2": "cheesy_fur_af21b42b95ec38c03db41b5a87aa938ebaedfcc6d3589f3983f64c285935b042.ext1.ext2",
		"foo":                  "foo_fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9",
	}
)
