/*
Split E-Mail Messages into text and attachments according to the parent package.
*/
package parser

import (
	"io"
	"io/ioutil"
	"github.com/byte-mug/mailsplit"
	"github.com/emersion/go-message/mail"
	netep "net/textproto"
)

var skipHeaders = map[string]bool{
	"Content-Type": true,
}

func Parse(aw mailsplit.AttachmentWriter, src io.Reader) (me *mailsplit.MailElement,err0 error) {
	me = new(mailsplit.MailElement)
	me.Header = make(map[string][]string)
	
	mr,e := mail.CreateReader(src)
	if e!=nil { return nil,e }
	
	for hf := mr.Header.Fields(); hf.Next(); {
		k := hf.Key()
		if skipHeaders[netep.CanonicalMIMEHeaderKey(k)] { continue }
		me.Header[k] = append(me.Header[k],hf.Value())
	}
	
	for {
		p,e := mr.NextPart()
		if e==io.EOF { break }
		switch v := p.Header.(type) {
		case *mail.InlineHeader:
			{
				var mt mailsplit.MailText
				mt.Format,_,_ = v.ContentType()
				b,e := ioutil.ReadAll(p.Body)
				if e!=nil && e!=io.EOF { return nil,e }
				mt.Body = string(b)
				me.Text = append(me.Text,mt)
			}
		case *mail.AttachmentHeader:
			ma := new(mailsplit.MailAttachment)
			ma.ContentType,_,_ = v.ContentType()
			ma.Filename,_ = v.Filename()
			if aw==nil {
				io.Copy(ioutil.Discard,p.Body)
			} else {
				w,e := aw.StoreAttachment(ma)
				if e!=nil { break }
				io.Copy(w,p.Body)
				c,_ := w.(io.Closer)
				if c!=nil { c.Close() }
			}
		}
	}
	return
}
