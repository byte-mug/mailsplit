package imapbe

import (
	"github.com/emersion/go-imap"
	"github.com/byte-mug/mailsplit"
	"strings"
	"encoding/hex"
	"crypto/md5"
)


func GetBodyStructure(me *mailsplit.MailElement, attachments []*mailsplit.MailAttachment, ext bool) (*imap.BodyStructure){
	bs := new(imap.BodyStructure)
	parts := make([]*imap.BodyStructure,0,3+len(attachments))
	
	bs.MIMEType = "multipart"
	bs.MIMESubType = "mixed"
	
	for _,te := range me.Text {
		ps := new(imap.BodyStructure)
		mt := append(strings.SplitN(te.Format,"/",2),"")
		ps.MIMEType = mt[0]
		ps.MIMESubType = mt[1]
		ps.Params = make(map[string]string)
		ps.Size = uint32(len(te.Body))
		ps.Lines = uint32(strings.Count(te.Body,"\n"))
		parts = append(parts,ps)
		if ext {
			ps.Extended = true
			ps.Disposition = "inline"
			ps.Language = me.Header["Content-Language"]
			sum := md5.Sum([]byte(te.Body))
			ps.MD5 = hex.EncodeToString(sum[:])
		}
	}
	for _,ma := range attachments {
		ps := new(imap.BodyStructure)
		mt := append(strings.SplitN(ma.ContentType,"/",2),"")
		ps.MIMEType = mt[0]
		ps.MIMESubType = mt[1]
		ps.Params = make(map[string]string)
		if ext {
			ps.Extended = true
			ps.Disposition = "attachment"
			bs.DispositionParams = map[string]string{ "name":ma.Filename }
			ps.Language = me.Header["Content-Language"]
		}
		parts = append(parts,ps)
	}
	
	bs.Parts = parts
	bs.Id = unwrap(me.Header["Content-Id"])
	bs.Description = unwrap(me.Header["Content-Description"])
	bs.Encoding = unwrap(me.Header["Content-Transfer-Encoding"])
	
	if ext {
		bs.Extended = true
		bs.Language = me.Header["Content-Language"]
		// bs.Disposition, bs.DispositionParams
		// bs.Language, bs.Location
		// bs.MD5
	}
	return bs
}
