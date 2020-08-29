package writer

import (
	"io"
	"github.com/byte-mug/mailsplit"
	"github.com/emersion/go-message/textproto"
	"github.com/emersion/go-message"
	
	// Encoding
	"github.com/emersion/go-textwrapper"
	"mime/quotedprintable"
	"encoding/base64"
	netep "net/textproto"
)


type MailAsText struct {
	Me *mailsplit.MailElement
	Attachments []mailsplit.MailAttachmentObject
	
	Dst io.Writer
	writeHdr func(textproto.Header) error
	fieldSet map[string]bool
	fieldState bool
}

func (mat *MailAsText) SetFieldSet(fs []string,not bool) {
	if len(fs)==0 {
		mat.fieldSet = nil
	} else {
		r := make(map[string]bool,len(fs))
		for _,s := range fs { r[netep.CanonicalMIMEHeaderKey(s)]=true }
		mat.fieldSet = r
	}
	mat.fieldState = !not
}

func (mat *MailAsText) pushDst() func() {
	tmp := mat.Dst
	tmp2 := mat.writeHdr
	tmp3 := mat.fieldSet
	tmp4 := mat.fieldState
	return func(){
		mat.Dst = tmp
		mat.writeHdr = tmp2
		mat.fieldSet = tmp3
		mat.fieldState = tmp4
	}
}
func (mat *MailAsText) doWriteHeader(hdr textproto.Header) error {
	for hf := hdr.Fields(); hf.Next(); {
		if mat.fieldSet[hf.Key()]!=mat.fieldState { hf.Del() }
	}
	if mat.writeHdr!=nil { return mat.writeHdr(hdr) }
	return textproto.WriteHeader(mat.Dst,hdr)
}
func (mat *MailAsText) EncodeText(i int,h bool,b bool) error {
	mt := &(mat.Me.Text[i])
	var usePlain = mt.Format=="text/plain"
	if h {
		var hdr message.Header
		hdr.SetContentType(mt.Format,nil)
		hdr.SetContentDisposition("inline",nil)
		if usePlain {
			hdr.Set("Content-Transfer-Encoding","8bit")
		} else {
			hdr.Set("Content-Transfer-Encoding","quoted-printable")
		}
		err := mat.doWriteHeader(hdr.Header)
		if err!=nil { return err }
	}
	if b {
		var w io.Writer
		cl := func() error { return nil }
		if usePlain {
			w = mat.Dst
		} else {
			t := quotedprintable.NewWriter(mat.Dst)
			w = t
			cl = t.Close
		}
		_,err := io.WriteString(w,mt.Body)
		if err!=nil { return err }
		err = cl()
		if err!=nil { return err }
	}
	return nil
}
func (mat *MailAsText) EncodeTexts(h bool,b bool) error {
	if len(mat.Me.Text)==1 {
		return mat.EncodeText(0,h,b)
	}
	tsep := "txt."+mat.Me.Seperator+".txt"
	if h {
		var hdr message.Header
		hdr.SetContentType("multipart/alternative",map[string]string{"boundary":tsep})
		err := mat.doWriteHeader(hdr.Header)
		if err!=nil { return err }
	}
	if b {
		defer mat.pushDst()()
		mat.SetFieldSet(nil,true)
		mpw := textproto.NewMultipartWriter(mat.Dst)
		mpw.SetBoundary(tsep)
		mat.writeHdr = func(hdr textproto.Header) (err error){
			mat.Dst,err = mpw.CreatePart(hdr)
			return
		}
		for i := range mat.Me.Text {
			mat.EncodeText(i,true,true)
		}
		mpw.Close()
	}
	return nil
}
func (mat *MailAsText) EncodeAtt(i int,h bool,b bool) error {
	att := mat.Attachments[i]
	ats := att.Att()
	if h {
		var hdr message.Header
		hdr.SetContentType(ats.ContentType,map[string]string{"name":ats.Filename})
		hdr.SetContentDisposition("attachment",map[string]string{"filename":ats.Filename})
		hdr.Set("Content-Transfer-Encoding","base64")
		err := mat.doWriteHeader(hdr.Header)
		if err!=nil { return err }
	}
	if b {
		r,err := att.Get()
		if err!=nil { return err }
		defer r.Close()
		w := base64.NewEncoder(base64.StdEncoding, textwrapper.NewRFC822(mat.Dst))
		_,err = io.Copy(w,r)
		if err!=nil { return err }
		err = w.Close()
		if err!=nil { return err }
	}
	return nil
}
func (mat *MailAsText) EncodeAll(h bool,b bool) error {
	if len(mat.Me.Text)==1 {
		return mat.EncodeText(0,h,b)
	}
	tsep := "pot."+mat.Me.Seperator+".top"
	if h {
		var hdr message.Header
		for k,vs := range mat.Me.Header {
			for _,v := range vs {
				hdr.Add(k,v)
			}
		}
		hdr.SetContentType("multipart/mixed",map[string]string{"boundary":tsep})
		err := mat.doWriteHeader(hdr.Header)
		if err!=nil { return err }
	}
	if b {
		defer mat.pushDst()()
		mat.SetFieldSet(nil,true)
		mpw := textproto.NewMultipartWriter(mat.Dst)
		mpw.SetBoundary(tsep)
		mat.writeHdr = func(hdr textproto.Header) (err error){
			mat.Dst,err = mpw.CreatePart(hdr)
			return
		}
		mat.EncodeTexts(true,true)
		for i := range mat.Attachments {
			mat.EncodeAtt(i,true,true)
		}
		mpw.Close()
	}
	return nil
}


