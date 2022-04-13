package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"

	"github.com/dumbboat/covid-tracker/model"
	"github.com/dumbboat/covid-tracker/util"
	"github.com/mxk/go-imap/imap"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"

	qprintable "github.com/sloonz/go-qprintable"
	"golang.org/x/net/html"
)

const DateFormat = "02-Jan-2006"

// Email is a simplified email struct containing the basic pieces of an email. If you want more info,
// it should all be available within the Message attribute.
type Email struct {
	Message *mail.Message

	From         *mail.Address   `json:"from"`
	To           []*mail.Address `json:"to"`
	InternalDate time.Time       `json:"internal_date"`
	Precedence   string          `json:"precedence"`
	Subject      string          `json:"subject"`
	HTML         []byte          `json:"html"`
	Text         []byte          `json:"text"`
	IsMultiPart  bool            `json:"is_multipart"`
	UID          uint32          `json:"uid"`
}

var (
	styleTag       = []byte("style")
	scriptTag      = []byte("script")
	headTag        = []byte("head")
	metaTag        = []byte("meta")
	doctypeTag     = []byte("doctype")
	shapeTag       = []byte("v:shape")
	imageDataTag   = []byte("v:imagedata")
	commentTag     = []byte("!")
	nonVisibleTags = [][]byte{
		styleTag,
		scriptTag,
		headTag,
		metaTag,
		doctypeTag,
		shapeTag,
		imageDataTag,
		commentTag,
	}
)

// ReorderByDate reorders your email by date
// order key word takes 'desc' or 'incr', the first one is for ordering by newest to olders, the other is the opposite.
// func ReorderByDate(mails []Email, order string) (orderedEmails []Email, err error) {
// 	if order != "desc" && order != "incr"{
// 		return mails,errors.New("order key word can only take 'desc' or 'incr'.")
// 	}
// 	if order == "desc"
// }

// GetAll will pull all emails from the email folder and return them as a list.
func GetAll(info model.Mailbox, markAsRead, delete bool) ([]Email, []error) {
	// call chan, put 'em in a list, return
	var emails []Email
	var errs []error
	responses, err := GenerateAll(info, markAsRead, delete)
	if err != nil {
		return emails, []error{err}
	}

	for resp := range responses {
		if resp.Err == nil {
			emails = append(emails, resp.Email)
		} else {
			errs = append(errs, resp.Err)
		}

	}

	return emails, errs
}

// GenerateAll will find all emails in the email folder and pass them along to the responses channel.
func GenerateAll(info model.Mailbox, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "ALL", nil, markAsRead, delete)
}

// GenerateCommand will find all emails in the email folder matching the IMAP command and pass them along to the responses channel.
func GenerateCommand(info model.Mailbox, IMAPCommand string, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, IMAPCommand, nil, markAsRead, delete)
}

// GetCommand will pull all emails that match the provided IMAP Command.
// Examples of IMAP Commands include TO/FROM/BCC, some examples are here http://www.marshallsoft.com/ImapSearch.htm
func GetCommand(info model.Mailbox, IMAPCommand string, markAsRead, delete bool) ([]Email, []error) {
	responses, err := GenerateCommand(info, IMAPCommand, markAsRead, delete)
	return responseToList(responses, err)
}

// GetUnread will find all unread emails in the folder and return them as a list.
func GetUnread(info model.Mailbox, markAsRead, delete bool) ([]Email, []error) {
	responses, err := GenerateUnread(info, markAsRead, delete)
	return responseToList(responses, err)
}

func responseToList(responses chan Response, err error) ([]Email, []error) {
	// call chan, put 'em in a list, return
	var emails []Email
	var errs []error
	if err != nil {
		return emails, []error{err}
	}

	for resp := range responses {
		if resp.Err == nil {
			emails = append(emails, resp.Email)
		} else {
			errs = append(errs, resp.Err)
		}

	}
	return emails, errs
}

// GenerateUnread will find all unread emails in the folder and pass them along to the responses channel.
func GenerateUnread(info model.Mailbox, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "UNSEEN", nil, markAsRead, delete)
}

// GetWithKeyMap will pull all emails that matches the given keymap.
func GetWithKeyMap(info model.Mailbox, keyMap map[string]interface{}, markAsRead, delete bool) ([]Email, []error) {
	var emails []Email
	var errs []error
	responses, err := GenerateWithKeyMap(info, keyMap, markAsRead, delete)
	if err != nil {
		return emails, []error{err}
	}

	for resp := range responses {
		if resp.Err == nil {
			emails = append(emails, resp.Email)
		} else {
			errs = append(errs, resp.Err)
		}

	}
	return emails, errs
}

// GetSince will pull all emails that have an internal date after the given time.
func GetSince(info model.Mailbox, since time.Time, markAsRead, delete bool) ([]Email, []error) {
	var emails []Email
	var errors []error
	responses, err := GenerateSince(info, since, markAsRead, delete)
	if err != nil {
		return emails, []error{err}
	}

	for resp := range responses {
		if resp.Err == nil {
			emails = append(emails, resp.Email)
			// return emails, resp.Err
		} else {
			errors = append(errors, resp.Err)
		}

	}

	return emails, errors
}

// GenerateSince will find all emails that have an internal date after the given time and pass them along to the
// responses channel.
func GenerateSince(info model.Mailbox, since time.Time, markAsRead, delete bool) (chan Response, error) {
	keyM := map[string]interface{}{
		"SINCE": since.Format(DateFormat),
	}
	return generateMail(info, "", keyM, markAsRead, delete)
}

// GenerateWithKeyMap will find all emails that matches the given keyMap
func GenerateWithKeyMap(info model.Mailbox, keyMap map[string]interface{}, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "", keyMap, markAsRead, delete)
}

// MarkAsUnread will set the UNSEEN flag on a supplied slice of UIDs
func MarkAsUnread(info model.Mailbox, uids []uint32) error {

	client, err := newIMAPClient(info)
	if err != nil {
		return err
	}
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	for _, u := range uids {
		err := alterEmail(client, u, "\\SEEN", false)
		if err != nil {
			return err //return on first failure
		}
	}
	return nil

}

// MarkAsUnread will set the UNSEEN flag on a supplied slice of UIDs
func MarkAsRead(info model.Mailbox, uids []uint32) error {

	client, err := newIMAPClient(info)
	if err != nil {
		return err
	}
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	for _, u := range uids {
		err := alterEmail(client, u, "\\SEEN", true)
		if err != nil {
			return err //return on first failure
		}
	}
	return nil

}

// DeleteEmails will delete emails from the supplied slice of UIDs
func DeleteEmails(info model.Mailbox, uids []uint32) error {

	client, err := newIMAPClient(info)
	if err != nil {
		return err
	}
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	for _, u := range uids {
		err := deleteEmail(client, u)
		if err != nil {
			return err //return on first failure
		}
	}
	return nil

}

// ValidateMailboxInfo attempts to login to the supplied IMAP account to ensure the info is correct
func ValidateMailboxInfo(info model.Mailbox) error {
	client, err := newIMAPClient(info)
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	return err
}

func VisibleText(body io.Reader) ([][]byte, error) {
	var (
		text [][]byte
		skip bool
		err  error
	)
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err = z.Err(); err == io.EOF {
				return text, nil
			}
			return text, err
		case html.TextToken:
			if !skip {
				tmp := bytes.TrimSpace(z.Text())
				if len(tmp) == 0 {
					continue
				}
				tagText := make([]byte, len(tmp))
				copy(tagText, tmp)
				text = append(text, tagText)
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			for _, nvTag := range nonVisibleTags {
				if bytes.Equal(tn, nvTag) {
					skip = (tt == html.StartTagToken)
					break
				}
			}
		}
	}
	return text, nil
}

// VisibleText will return any visible text from an HTML
// email body.
func (e *Email) VisibleText() ([][]byte, error) {
	// if theres no HTML, just return text
	if len(e.HTML) == 0 {
		return [][]byte{e.Text}, nil
	}
	return VisibleText(bytes.NewReader(e.HTML))
}

// String is to spit out a somewhat pretty version of the email.
func (e *Email) String() string {
	return fmt.Sprintf(`
----------------------------
From:           %s
To:             %s
Internal Date:  %s
Precedence:     %s
Subject:        %s
HTML:           %s
Text:           %s
----------------------------
`,
		e.From,
		e.To,
		e.InternalDate,
		e.Precedence,
		e.Subject,
		string(e.HTML),
		string(e.Text),
	)
}

// Response is a helper struct to wrap the email responses and possible errors.
type Response struct {
	Email Email
	Err   error
}

// newIMAPClient will initiate a new IMAP connection with the given creds.
func newIMAPClient(info model.Mailbox) (*imap.Client, error) {
	var client *imap.Client
	var err error
	if info.TLS {
		config := new(tls.Config)
		config.InsecureSkipVerify = info.InsecureSkipVerify
		client, err = imap.DialTLS(info.Host, config)
		if err != nil {
			return client, err
		}
	} else {
		client, err = imap.Dial(info.Host)
		if err != nil {
			return client, err
		}
	}

	_, err = client.Login(info.User, info.Pwd)
	if err != nil {
		return client, err
	}

	_, err = imap.Wait(client.Select(info.Folder, info.ReadOnly))
	if err != nil {
		return client, err
	}

	return client, nil
}

// findEmails will run a find the UIDs of any emails that match the search.:
func findEmails(client *imap.Client, search string, keyMap map[string]interface{}) (*imap.Command, error) {
	fmt.Println(keyMap)
	var specs []imap.Field
	if len(search) > 0 {
		specs = append(specs, search)
	}
	for k, v := range keyMap {
		if _, ok := allowedKeys[k]; ok {
			specs = append(specs, k, v)
		}
	}
	// get headers and UID for UnSeen message in src inbox...
	cmd, err := imap.Wait(client.UIDSearch(specs...))
	if err != nil {
		return &imap.Command{}, fmt.Errorf("uid search failed: %s", err)
	}
	return cmd, nil
}

var GenerateBufferSize = 100

func generateMail(info model.Mailbox, search string, keyMap map[string]interface{}, markAsRead, delete bool) (chan Response, error) {
	responses := make(chan Response, GenerateBufferSize)
	client, err := newIMAPClient(info)
	if err != nil {
		close(responses)
		return responses, fmt.Errorf("uid search failed: %s", err)
	}
	go func() {
		defer func() {
			client.Close(true)
			client.Logout(30 * time.Second)
			close(responses)
		}()

		var cmd *imap.Command
		// find all the UIDs
		cmd, err = findEmails(client, search, keyMap)
		if err != nil {
			responses <- Response{Err: err}
			return
		}
		// gotta fetch 'em all

		getEmails(client, cmd, markAsRead, delete, responses)

	}()
	return responses, nil
}

func getEmails(client *imap.Client, cmd *imap.Command, markAsRead, delete bool, responses chan Response) {
	seq := &imap.SeqSet{}
	msgCount := 0
	for _, rsp := range cmd.Data {
		for _, uid := range rsp.SearchResults() {
			msgCount++
			seq.AddNum(uid)
		}
	}
	// nothing to request?! why you even callin me, foolio?
	if seq.Empty() {
		return
	}

	fCmd, err := imap.Wait(client.UIDFetch(seq, "INTERNALDATE", "BODY[]", "UID", "RFC822.HEADER"))
	if err != nil {
		responses <- Response{Err: fmt.Errorf("unable to perform uid fetch: %s", err)}
		return
	}

	var email Email
	for _, msgData := range fCmd.Data {
		msgFields := msgData.MessageInfo().Attrs

		// make sure is a legit response before we attempt to parse it
		// deal with unsolicited FETCH responses containing only flags
		// I'm lookin' at YOU, Gmail!
		// http://mailman13.u.washington.edu/pipermail/imap-protocol/2014-October/002355.html
		// http://stackoverflow.com/questions/26262472/gmail-imap-is-sometimes-returning-bad-results-for-fetch
		if _, ok := msgFields["RFC822.HEADER"]; !ok {
			continue
		}

		email, err = NewEmail(msgFields)
		if err != nil {
			responses <- Response{Err: fmt.Errorf("unable to parse email: %s", err)}
			continue
		}

		responses <- Response{Email: email}

		if markAsRead {
			err = addSeen(client, imap.AsNumber(msgFields["UID"]))
			if err != nil {
				responses <- Response{Err: fmt.Errorf("unable to add seen flag: %s", err)}
				continue
			}

		} else {
			err = removeSeen(client, imap.AsNumber(msgFields["UID"]))
			if err != nil {
				responses <- Response{Err: fmt.Errorf("unable to remove seen flag: %s", err)}
				continue
			}
		}

		if delete {
			err = deleteEmail(client, imap.AsNumber(msgFields["UID"]))
			if err != nil {
				responses <- Response{Err: fmt.Errorf("unable to delete email: %s", err)}
				continue
			}
		}
	}
}

func deleteEmail(client *imap.Client, UID uint32) error {
	return alterEmail(client, UID, "\\DELETED", true)
}

func removeSeen(client *imap.Client, UID uint32) error {
	return alterEmail(client, UID, "\\SEEN", false)
}

func addSeen(client *imap.Client, UID uint32) error {
	return alterEmail(client, UID, "\\SEEN", true)
}

func alterEmail(client *imap.Client, UID uint32, flag string, plus bool) error {
	flg := "-FLAGS"
	if plus {
		flg = "+FLAGS"
	}
	fSeq := &imap.SeqSet{}
	fSeq.AddNum(UID)
	_, err := imap.Wait(client.UIDStore(fSeq, flg, flag))
	if err != nil {
		return err
	}

	return nil
}

func hasEncoding(word string) bool {
	return strings.Contains(word, "=?") && strings.Contains(word, "?=")
}

func isEncodedWord(word string) bool {
	return strings.HasPrefix(word, "=?") && strings.HasSuffix(word, "?=") && strings.Count(word, "?") == 4
}

func parseSubject(subject string) string {
	if !hasEncoding(subject) {
		return subject
	}

	dec := mime.WordDecoder{}
	convertedSubject, err := util.GbkToUtf8([]byte(subject))
	if err != nil {
		panic(err)
	}
	sub, _ := dec.DecodeHeader(string(convertedSubject))
	return sub
}

// NewEmail will parse an imap.FieldMap into an Email. This
// will expect the message to container the internaldate and the body with
// all headers included.
func NewEmail(msgFields imap.FieldMap) (Email, error) {
	var email Email
	// parse the header
	var message bytes.Buffer
	convertedHeader, err := util.GbkToUtf8(imap.AsBytes(msgFields["RFC822.HEADER"]))
	if err != nil {
		panic(err)
	}
	message.Write(convertedHeader)
	message.Write([]byte("\n\n"))
	rawBody := imap.AsBytes(msgFields["BODY[]"])
	converted, err := util.GbkToUtf8(rawBody)
	if err != nil {
		panic(err)
	}
	message.Write(converted)
	msg, err := mail.ReadMessage(&message)
	if err != nil {
		return email, fmt.Errorf("unable to read header: %s", err)
	}

	from, err := mail.ParseAddress("Unkown <unknown@example.com>")
	if err != nil {
		return email, fmt.Errorf("unable to parse from address: %s", err)
	}

	to, err := mail.ParseAddressList(msg.Header.Get("To"))
	if err != nil {
		to = []*mail.Address{}
	}

	email = Email{
		Message:      msg,
		InternalDate: imap.AsDateTime(msgFields["INTERNALDATE"]),
		Precedence:   msg.Header.Get("Precedence"),
		From:         from,
		To:           to,
		Subject:      parseSubject(msg.Header.Get("Subject")),
		UID:          imap.AsNumber(msgFields["UID"]),
	}

	// chunk the body up into simple chunks
	email.HTML, email.Text, email.IsMultiPart, err = parseBody(msg.Header, rawBody)
	return email, err
}

var headerSplitter = []byte("\r\n\r\n")

// parseBody will accept a a raw body, break it into all its parts and then convert the
// message to UTF-8 from whatever charset it may have.
func parseBody(header mail.Header, body []byte) (html []byte, text []byte, isMultipart bool, err error) {
	var mediaType string
	var params map[string]string
	mediaType, params, err = mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		isMultipart = true
		mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			slurp, err := ioutil.ReadAll(p)
			if err != nil {
				// error and no results to use
				if len(slurp) == 0 {
					break
				}
			}

			partMediaType, partParams, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
			if err != nil {
				break
			}

			var htmlT, textT []byte
			htmlT, textT, err = parsePart(partMediaType, partParams["charset"], p.Header.Get("Content-Transfer-Encoding"), slurp)
			if len(htmlT) > 0 {
				html = htmlT
			} else {
				text = textT
			}
		}
	} else {

		splitBody := bytes.SplitN(body, headerSplitter, 2)
		if len(splitBody) < 2 {
			err = errors.New("unexpected email format. (single part and no \\r\\n\\r\\n separating headers/body")
			return
		}

		body = splitBody[1]
		html, text, err = parsePart(mediaType, params["charset"], header.Get("Content-Transfer-Encoding"), body)
	}
	return
}

func parsePart(mediaType, charsetStr, encoding string, part []byte) (html, text []byte, err error) {
	// deal with charset
	if strings.ToLower(charsetStr) == "iso-8859-1" {
		var cr io.Reader
		cr, err = charset.NewReader("latin1", bytes.NewReader(part))
		if err != nil {
			return
		}

		part, err = ioutil.ReadAll(cr)
		if err != nil {
			return
		}
	}

	// deal with encoding
	var body []byte
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		dec := qprintable.NewDecoder(qprintable.WindowsTextEncoding, bytes.NewReader(part))
		body, err = ioutil.ReadAll(dec)
		if err != nil {
			return
		}
	case "base64":
		decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(part))
		body, err = ioutil.ReadAll(decoder)
		if err != nil {
			return
		}
	default:
		body = part
	}

	// deal with media type
	mediaType = strings.ToLower(mediaType)
	switch {
	case strings.Contains(mediaType, "text/html"):
		html = body
	case strings.Contains(mediaType, "text/plain"):
		text = body
	}
	return
}
