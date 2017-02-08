package crawler

import (
	"testing"
	"github.com/therahulprasad/spiderman/crawler/db"
)

func TestAll(t *testing.T) {
	link := "https://www.sampada.net/ಮಲೆನಾಡಿನ-ಮಾಳ-ಕಾವಲು-ಕೊನೆಗೆ-ಕಂಬಳ-ಭಾಗ-1/47442"

	Setup("/home/rahul/Development/Go/src/github.com/therahulprasad/spiderman/config.json")
	_, err := db.Push(link, 0)
	if err != nil {
		t.Error(err.Error())
	}

	node, err := db.Pop()
	if err != nil {
		t.Error(err.Error())
	}

	if node.Link != link {
		t.Error("Link mismatch")
	}

	err = db.Update(node.Id, 1, db.Success)
	if err != nil {
		t.Error(err.Error())
	}

	page_processor(link)

}