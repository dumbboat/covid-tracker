package mail

var allowedKeys = map[string]interface{}{
	"ALL":        nil, //All messages in the mailbox; the default initial key for ANDing.
	"ANSWERED":   nil, //Messages with the \Answered flag set.
	"BCC":        nil, //Messages that contain the specified string in the envelope structure's BCC field.
	"BEFORE":     nil, //Messages whose internal date is earlier than the specified date.
	"BODY":       nil, //Messages that contain the specified string in the body of the message.
	"CC":         nil, //Messages that contain the specified string in the envelope structure's CC field.
	"DELETED":    nil, //Messages with the \Deleted flag set.
	"DRAFT":      nil, //Messages with the \Draft flag set.
	"FLAGGED":    nil, //Messages with the \Flagged flag set.
	"FROM":       nil, //Messages that contain the specified string in the envelope structure's FROM field.
	"HEADER":     nil, //'field-name' 'string' Messages that have a header with the specified field-name (as defined in [RFC-822]) and that contains the specified string in the [RFC-822] field-body.
	"KEYWORD":    nil, //Messages with the specified keyword set.
	"LARGER":     nil, //Messages with an RFC822.SIZE larger than the specified number of octets.
	"NEW":        nil, //Messages that have the \Recent flag set but not the \Seen flag. This is functionally equivalent to "(RECENT UNSEEN)".
	"NOT":        nil, //Messages that do not match the specified search key.
	"OLD":        nil, //Messages that do not have the \Recent flag set. This is functionally equivalent to "NOT RECENT" (as opposed to "NOT NEW").
	"ON":         nil, //Messages whose internal date is within the specified date.
	"OR":         nil, //'search-key1' 'search-key2' Messages that match either search key.
	"RECENT":     nil, //Messages that have the \Recent flag set.
	"SEEN":       nil, //Messages that have the \Seen flag set.
	"SENTBEFORE": nil, //Messages whose [RFC-822] Date: header is earlier than the specified date.
	"SENTON":     nil, //Messages whose [RFC-822] Date: header is within the specified date.
	"SENTSINCE":  nil, //Messages whose [RFC-822] Date: header is within or later than the specified date.
	"SINCE":      nil, //Messages whose internal date is within or later than the specified date.
	"SMALLER":    nil, //Messages with an RFC822.SIZE smaller than the specified number of octets.
	"SUBJECT":    nil, //Messages that contain the specified string in the envelope structure's SUBJECT field.
	"TEXT":       nil, //Messages that contain the specified string in the header or body of the message.
	"TO":         nil, //Messages that contain the specified string in the envelope structure's TO field.
	"UID":        nil, //Messages with unique identifiers corresponding to the specified unique identifier set.
	"UNANSWERED": nil, //Messages that do not have the \Answered flag set.
	"UNDELETED":  nil, //Messages that do not have the \Deleted flag set.
	"UNDRAFT":    nil, //Messages that do not have the \Draft flag set.
	"UNFLAGGED	": nil, //Messages that do not have the \Flagged flag set.
	"UNKEYWORD": nil, //Messages that do not have the specified keyword set.
	"UNSEEN":    nil, //Messages that do not have the \Seen flag set.
}
