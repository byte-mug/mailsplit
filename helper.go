package mailsplit

import (
	"io"
)

type AttachmentWriter interface{
	StoreAttachment(h *MailAttachment) (io.Writer,error)
}

