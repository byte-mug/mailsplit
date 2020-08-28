/*
IMAP Backend Utility.
*/
package imapbe

import (
	"net/mail"
	"github.com/emersion/go-imap"
	"github.com/byte-mug/mailsplit"
	"strings"
)

func unwrap(s []string) string {
	if len(s)==0 { return "" }
	return s[0]
}

func parseAddress(sl []string) (ret []*imap.Address) {
	ret = make([]*imap.Address,0,8)
	for _,s := range sl {
		pal,err := mail.ParseAddressList(s)
		if err!=nil { break }
		for _,addr := range pal {
			sub := append(strings.SplitN(addr.Address,"@",2),"")
			ret = append(ret,&imap.Address{
				PersonalName: addr.Name,
				MailboxName: sub[0],
				HostName: sub[1],
			})
		}
	}
	return
}

func GetEnvelope(me *mailsplit.MailElement) (*imap.Envelope){
	env := new(imap.Envelope)
	
	// Filling Fields in chronilogical Order
	env.Date,_ = mail.ParseDate(unwrap(me.Header["Date"]))
	env.Subject = unwrap(me.Header["Subject"])
	
	env.From = parseAddress(me.Header["From"])
	env.Sender = parseAddress(me.Header["Sender"])
	env.ReplyTo = parseAddress(me.Header["Reply-To"])
	
	env.To = parseAddress(me.Header["To"])
	env.Cc = parseAddress(me.Header["Cc"])
	env.Bcc = parseAddress(me.Header["Bcc"])
	env.InReplyTo = unwrap(me.Header["In-Reply-To"])
	env.MessageId = unwrap(me.Header["Message-Id"])
	
	if len(env.Sender) == 0 {
		env.Sender = env.From
	}
	if len(env.ReplyTo) == 0 {
		env.ReplyTo = env.From
	}
	return env
}

