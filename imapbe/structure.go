package imapbe

import (
	"github.com/emersion/go-imap"
	"github.com/byte-mug/mailsplit"
	"strings"
	"encoding/hex"
	"crypto/md5"
	
	"mime/quotedprintable"
	"bytes"
)

func convert(bb []byte,plain bool) []byte {
	if plain { return bb }
	b := new(bytes.Buffer)
	w := quotedprintable.NewWriter(b)
	w.Write(bb)
	w.Close()
	return b.Bytes()
}

func GetBodyStructure(me *mailsplit.MailElement, attachments []mailsplit.MailAttachmentObject, ext bool) (*imap.BodyStructure){
	bs := new(imap.BodyStructure)
	parts := make([]*imap.BodyStructure,0,1+len(attachments))
	texts := make([]*imap.BodyStructure,0,len(me.Text))
	
	bs.MIMEType = "multipart"
	bs.MIMESubType = "mixed"
	bs.Params = map[string]string {"boundary":"pot."+me.Seperator+".top"}
	
	for _,te := range me.Text {
		var usePlain = te.Format=="text/plain"
		ps := new(imap.BodyStructure)
		mt := append(strings.SplitN(te.Format,"/",2),"")
		ps.MIMEType = mt[0]
		ps.MIMESubType = mt[1]
		ps.Params = make(map[string]string)
		bb := convert([]byte(te.Body),usePlain)
		
		ps.Size = uint32(len(bb))
		ps.Lines = uint32(strings.Count(te.Body,"\n"))
		if usePlain {
			ps.Encoding = "8bit"
		} else {
			ps.Encoding = "quoted-printable"
		}
		texts = append(texts,ps)
		if ext {
			ps.Extended = true
			ps.Disposition = "inline"
			ps.Language = me.Header["Content-Language"]
			sum := md5.Sum(bb)
			ps.MD5 = hex.EncodeToString(sum[:])
		}
	}
	if len(texts)==1 {
		parts = append(parts,texts[0])
	} else {
		ps := new(imap.BodyStructure)
		ps.MIMEType = "multipart"
		ps.MIMESubType = "alternative"
		ps.Params = map[string]string {"boundary":"txt."+me.Seperator+".txt"}
		ps.Parts = texts
		parts = append(parts,ps)
		if ext {
			ps.Extended = true
			ps.Disposition = "inline"
			ps.Language = me.Header["Content-Language"]
		}
	}
	for _,mao := range attachments {
		ma := mao.Att()
		ps := new(imap.BodyStructure)
		mt := append(strings.SplitN(ma.ContentType,"/",2),"")
		ps.MIMEType = mt[0]
		ps.MIMESubType = mt[1]
		ps.Encoding = "base64"
		if ma.Filename!="" {
			ps.Params = map[string]string{ "name": ma.Filename }
		}
		if l,err := mao.Length(); err==nil {
			/*
			Approximation:
			 base64 lines are 76 Characters long (= 19*4)
			 each line is terminated with "\r\n" (2 bytes)
			 4 bytes are generated out of 3 bytes binary data.
			*/
			m := (l+2)/3 // Number of 4 byte Base64 words (encodes 3 bytes from SRC)
			lb := (m+18)/19 // Chop into 19 Word Lines (19*4 = 76 bytes)
			total := (m*4)+(lb*2) // Compute the total number of bytes m base64 words + lb linebreaks.
			if total>0xFFFFFFFF { total = 0xFFFFFFFF } // Prevent Buffer overflow.
			ps.Size = uint32(total) // Set size of body
			ps.Lines = uint32(lb) // Set line count of body
		}
		if ext {
			ps.Extended = true
			ps.Disposition = "attachment"
			if ma.Filename!="" {
				ps.DispositionParams = map[string]string{ "filename": ma.Filename }
			}
			ps.Language = me.Header["Content-Language"]
		}
		parts = append(parts,ps)
	}
	
	bs.Parts = parts
	bs.Id = unwrap(me.Header["Content-Id"])
	bs.Description = unwrap(me.Header["Content-Description"])
	//bs.Encoding = unwrap(me.Header["Content-Transfer-Encoding"])
	
	if ext {
		bs.Extended = true
		bs.Language = me.Header["Content-Language"]
		// bs.Disposition, bs.DispositionParams
		// bs.Language, bs.Location
		// bs.MD5
	}
	return bs
}
