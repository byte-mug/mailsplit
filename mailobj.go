/*
Offers a compact representation for E-Mails, with Attachments Seperated.
That makes it easy to split E-Mail Messages into text and attachments,
store them and recombine them in a consistent way.

All info-structures can be serialized as JSON to allow efficient storage of them.
*/
package mailsplit

import "io"

type MailAttachment struct {
	ContentType string
	Filename string
}
type MailText struct {
	Format string
	Body string
}

type MailElement struct {
	Header map[string][]string
	Text []MailText
	Seperator string
}

type MailAttachmentObject interface{
	Att() *MailAttachment
	Get() (io.ReadCloser,error)
	// Estimate the length.
	Length() (int64,error)
}
