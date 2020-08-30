package imapbe

import (
	"net/mail"
	"github.com/emersion/go-imap"
	"github.com/byte-mug/mailsplit"
	"strings"
	"time"
	"regexp"
)

// This is the signature of a Comment
var remCommentar = regexp.MustCompile(`[ \t]+\(.*\)$`)

type ImapMailboxMsg struct {
	SeqNum, Uid uint32
	Date time.Time
	
	Keywords []string
}

type matchMailElement struct {
	me *mailsplit.MailElement
	meta *ImapMailboxMsg
	
	kwm map[string]bool
	sent time.Time
}
func (mme *matchMailElement) kwmap() map[string]bool {
	if mme.kwm==nil {
		mme.kwm = make(map[string]bool)
		for _,k := range mme.meta.Keywords { mme.kwm[k] = true }
	}
	return mme.kwm
}
func (mme *matchMailElement) getsent() (t time.Time,e error) {
	t = mme.sent
	if t.IsZero() {
		// Date headers sometimes contain a Comment.
		date := remCommentar.ReplaceAllString(unwrap(mme.me.Header["Date"]),"")
		t,e = mail.ParseDate(date)
		mme.sent = t
	}
	return
}

func errand(e1, e2 error) error {
	if e1!=nil { return e1 }
	return e2
}

type matcher struct {
	rxm map[string]*regexp.Regexp
	
	mmebuf matchMailElement
	root *imap.SearchCriteria
}
func (m *matcher) regex(str string) *regexp.Regexp {
	if m.rxm==nil { m.rxm = make(map[string]*regexp.Regexp,8) }
	if re := m.rxm[str]; re!=nil { return re }
	lstr := strings.ToLower(str)
	if re := m.rxm[lstr]; re!=nil {
		m.rxm[str] = re
		return re
	}
	rx := regexp.QuoteMeta(lstr)
	re := regexp.MustCompile(`(?i(`+rx+`))`)
	m.rxm[lstr] = re
	m.rxm[str] = re
	return re
}
func (m *matcher) matchSingle(c *imap.SearchCriteria, mme *matchMailElement) (bool,error) {
	if c.SeqNum!=nil {
		if !c.SeqNum.Contains(mme.meta.SeqNum) { return false, nil }
	}
	if c.Uid!=nil {
		if !c.Uid.Contains(mme.meta.Uid) { return false, nil }
	}
	date := mme.meta.Date
	if c.Since.After(date) { return false, nil }
	if !(c.Before.IsZero() || c.Before.After(date)) { return false, nil }
	if len(c.WithFlags)!=0 || len(c.WithoutFlags)!=0 {
		kwm := mme.kwmap()
		for _,f := range c.WithFlags {
			if !kwm[f] { return false, nil }
		}
		for _,f := range c.WithoutFlags {
			if kwm[f] { return false, nil }
		}
	}
	if !(c.SentBefore.IsZero() && c.SentSince.IsZero()) {
		sent,err := mme.getsent()
		if err!=nil { return false, err }
		if c.SentSince.After(sent) { return false, nil }
		if !(c.SentBefore.IsZero() || c.SentBefore.After(sent)) { return false, nil }
	}
	for k,vs := range c.Header {
		elems,has := mme.me.Header[k]
		for _,v := range vs {
			if !has { return false, nil }
			if v=="" { continue }
			fnd := false
			rx := m.regex(v)
			for _,elem := range elems {
				if !rx.MatchString(elem) { continue }
				fnd = true
			}
			if !fnd { return false, nil }
		}
	}
	
	for _, body := range c.Body {
		rx := m.regex(body)
		found := false
		for _,t := range mme.me.Text {
			found = rx.MatchString(t.Body)
			if found { break }
		}
		if !found { return false, nil }
	}
	for _, body := range c.Text {
		rx := m.regex(body)
		found := false
		for _,t := range mme.me.Text {
			found = rx.MatchString(t.Body)
			if found { break }
		}
		if !found {
			for k,vs := range mme.me.Header {
				for _,v := range vs {
					found = rx.MatchString(k+": "+v)
					if found { break }
				}
				if found { break }
			}
		}
		if !found { return false, nil }
	}
	
	if !(c.Larger==0 && c.Smaller==0) {
		// XXXMFG: we can't estimate the size correctly.
		n := 0
		for _,t := range mme.me.Text {
			n += len(t.Body) + 128
		}
		n32 := uint32(n)
		if n32 <= c.Larger { return false, nil }
		if c.Smaller > 0 && c.Smaller <= n32 { return false, nil }
	}
	return true, nil
}
func (m *matcher) match(c *imap.SearchCriteria, mme *matchMailElement) (b bool,e error) {
	var nb bool
	var e2 error
	b,e = m.matchSingle(c,mme)
	if e!=nil || !b { return }
	
	b = false
	// From now on, return would return FALSE
	for _, nc := range c.Not {
		nb,e = m.match(nc,mme)
		if e!=nil || nb { return }
	}
	
	// In this loop, the return value will be overwritten.
	for _, oc := range c.Or {
		nb,e = m.match(oc[0],mme)
		b,e2 = m.match(oc[1],mme)
		e = errand(e,e2)
		if e!=nil || !(b||nb) { return }
	}
	
	// Set the return value back to true
	b = true
	return
}

func (m *matcher) Match(me *mailsplit.MailElement, meta *ImapMailboxMsg) (bool,error) {
	mme := &m.mmebuf
	*mme = matchMailElement{me:me, meta:meta}
	
	return m.match(m.root,mme)
}

type MailMatcher interface {
	Match(me *mailsplit.MailElement, meta *ImapMailboxMsg) (bool,error)
}
func NewMailMatcher(c *imap.SearchCriteria) MailMatcher {
	return &matcher{root:c }
}

