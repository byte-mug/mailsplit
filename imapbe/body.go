package imapbe

import (
	"github.com/emersion/go-imap"
	"github.com/byte-mug/mailsplit"
	//"github.com/emersion/go-message/mail"
	"github.com/byte-mug/mailsplit/writer"
	"bytes"
	"errors"
)

var errNoSuchPart = errors.New("imapbe: no such message body part")

func GetBodySection(me *mailsplit.MailElement, attachments []mailsplit.MailAttachmentObject, section *imap.BodySectionName) (imap.Literal, error) {
	var err error
	buf := new(bytes.Buffer)
	mat := new(writer.MailAsText)
	mat.Me = me
	mat.Attachments = attachments
	mat.Dst = buf
	
	h := true
	b := true
	
	switch section.Specifier {
	case imap.EntireSpecifier:
		h = len(section.Path)==0
	case imap.HeaderSpecifier: b = false
	case imap.TextSpecifier: h = false
	}
	
	if len(section.Fields)!=0 {
		mat.SetFieldSet(section.Fields,section.NotFields)
	}
	
	switch len(section.Path) {
	case 0:
		// Return Everything
		err = mat.EncodeAll(h,b)
	case 1:
		if section.Path[0]<1 { return nil, errNoSuchPart }
		if section.Path[0]==1 {
			// Return the multipart/alternative part.
			err = mat.EncodeTexts(h,b)
			if err!=nil { return nil, err }
			return buf,nil
		}
		// Return a particlar attachment
		aidx := section.Path[0]-2
		if aidx>=len(attachments) { return nil, errNoSuchPart }
		err = mat.EncodeAtt(aidx,h,b)
	case 2:
		// Return a particular inline part
		if section.Path[0]!=1 { return nil, errNoSuchPart }
		if section.Path[1]<1 { return nil, errNoSuchPart }
		tidx := section.Path[1]-1
		err = mat.EncodeText(tidx,h,b)
	default:
		return nil, errNoSuchPart
	}
	if err!=nil { return nil, err }
	
	if section.Partial != nil {
		return bytes.NewReader(section.ExtractPartial(buf.Bytes())), nil
	}
	return buf, nil
}
