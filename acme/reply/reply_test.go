package main

import (
	"testing"
	"net/mail"
)

func TestSubjectRewriting(t *testing.T) {
	if got, want := rewriteSubject("Piddle"), "Subject: Re:  Piddle\n"; got != want {
		t.Errorf("%s: got: %q, want: %q\n", "rewriteSubject", got, want)
	}

	if got, want := rewriteSubject("Re: fantastic"), "Subject: Re: fantastic\n"; got != want {
		t.Errorf("%s: got: %q, want: %q\n", "rewriteSubject", got, want)
	}
}

func TestFromInverting(t *testing.T) {
	astring := `"Eh Man" <a@foo.com>`
	a, err := mail.ParseAddress(astring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", astring, err)
	}

	bstring := `b@foo.com`
	b, err := mail.ParseAddress(bstring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", bstring, err)
	}

	cstring := `"C Lang" <c@prog.com>`
	c, err := mail.ParseAddress(cstring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", cstring, err)
	}

	dstring := `"D Light" <delite@fuddle.cn>`
	d, err := mail.ParseAddress(dstring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", dstring, err)
	}

	// TODO(rjk): Test flakes because I need to sort the result of invertedFrom.
	if got, want := stringifyAddressList(invertedFrom(
			[]*mail.Address{a},
			[]*mail.Address{d},
			d)), "\"Eh Man\" <a@foo.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList(invertedFrom(
			[]*mail.Address{a},
			[]*mail.Address{c,d},
			d)), "\"Eh Man\" <a@foo.com>, \"C Lang\" <c@prog.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList(invertedFrom(
			[]*mail.Address{a},
			[]*mail.Address{b, c,d},
			d)), "\"Eh Man\" <a@foo.com>, <b@foo.com>, \"C Lang\" <c@prog.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList(invertedFrom(
			nil,
			[]*mail.Address{b, c,d},
			d)), "<b@foo.com>, \"C Lang\" <c@prog.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestStringifyAddressList(t *testing.T) {
	astring := `"Eh Man" <a@foo.com>`
	a, err := mail.ParseAddress(astring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", astring, err)
	}

	bstring := `b@foo.com`
	b, err := mail.ParseAddress(bstring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", bstring, err)
	}

	cstring := `"C Lang" <c@prog.com>`
	c, err := mail.ParseAddress(cstring)
	if err != nil {
		t.Errorf("can't make example email from %q: %v", cstring, err)
	}

	if got, want := stringifyAddressList([]*mail.Address{a}), "\"Eh Man\" <a@foo.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList([]*mail.Address{a, b}), 
		"\"Eh Man\" <a@foo.com>, <b@foo.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList([]*mail.Address{a, b, c}), 
		"\"Eh Man\" <a@foo.com>, <b@foo.com>, \"C Lang\" <c@prog.com>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}

	if got, want := stringifyAddressList([]*mail.Address{}), ""; got != want {
		t.Errorf("got %q want %q", got, want)
	}

}
