package ofbx

import (
	"encoding/binary"
	"bufio"
)

type Header struct {
	magic [21]int
	reserved [2]int
	version int
}

type Cursor *bufio.Reader

const(
	UINT8_BYTES = 1
	UINT32_BYTES = 4
	HEADER_BYTES = (21 + 2 + 1) * 4
)

func (c *Cursor) readShortString() (string, error) {
	length, err := c.ReadByte()
	if err != nil {
		return nil, err
	}
	byt := make([]byte, int(length))
	_, err = c.Read(byt)
	if err != nil {
		return nil, err
	} 
	return string(byt), nil
}
func (c *Cursor ) readLongString() (string, error) {
	var length uint32
	err := binary.Read(c, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}
	byt := make([]byte, length)
	_, err = c.Read(byt)
	if err != nil {
		return "", err
	} 
	return string(byt), nil
}

func (c *Cursor) isEndLine() bool {
	by, err := c.Peek(1)
	if err != nil {
		fmt.Println(err)
	}
	return by[0] == '\n'
}
//TODO: Make a isspace for bytes not unicode
func (c *Cursor) skipInsignificantWhitespaces() error {
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsSpace(by[0]) && by[0] != '\n' {
			continue
		}
		c.UnreadRune()
		break
	}
}

func (c *Cursor) skipLine(){
	c.ReadLine()
}

func (c *Cursor) skipWhitespaces(){
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsSpace(by[0]) {
			continue
		}
		c.UnreadRune()
		break
	}
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if by[0] == ';' {
			c.skipLine()
			continue
		}
		c.UnreadRune()
		break
	}
}

func isTextTokenChar(c rune) {
	return unicode.IsDigit(c) || unicode.IsLetter(c) ||  c == '_'
}


func (c *Cursor) readTextToken() DataView {
	out := bytes.NewBuffer([]byte{})
	for {
		r, _, err := c.ReadRune()
		if err != nil {
			fmt.Println(err)
		}
		if isTextTokenChar(r) {
			out.WriteRune(r)
			continue
		}
		c.UnreadRune()
	}
	return DataView(out)
}


func (c *Cursor) readElementOffset(version uint16) (uint64, error) {
	if version >= 7500{
		var i uint64 
		err := binary.Read(c, binary.BigEndian, &i)
		return i, err
	}
	var i uint32 
	err := binary.Read(c, binary.BigEndian, &i)
	return uint64(i), err
}

func (c *Curesor) readBytes(len int) []byte{
	tempArr := make([]byte, len)
	_, err := c.Read(tempArr)
	return tempArr
}

func (c *Cursor) readProperty() *Property{
	if c.cur > len(c.data){
		return errors.NewError("Reading Past End")
	}
		prop := Property{}
		typ, _, err := c.ReadRune()
		if err !=nil{
			fmt.Println(err)
		}
		prop.typ = typ
		switch(prop.typ){
		case 'S':
			prop.value, err = c.readLongString()
		case 'Y': 
			prop.value = c.readBytes(2)
		case 'C': 
		prop.value =c.readBytes(1)
		case 'I': 
		prop.value =c.readBytes(4)
		case 'F': 
		prop.value =c.readBytes(4)
		case 'D': 
		prop.value =c.readBytes(8)
		case 'L': 
		prop.value =c.readBytes(8)
		case 'R':
		 	tmp := c.readBytes(4)
			length :=  binary.BigEndian.Uint32(tmp)
			prop.value = append(tmp,c.readBytes(length)...)
		case 'b'||'f'||'d'||'l'||'i':
			temp := c.readBytes(8)
			tempArr := c.readBytes(4)
			length :=  binary.BigEndian.Uint32(tempArr)
			prop.value = append(append(temp, tempArr...),c.readBytes(length)...)
		default:
			errors.NewError("Did not know this property")
		}
		if err != nil{
			fmt.Println(err)
		}
		return &prop
}

func (c *Cursor) readElement(version int){
	initial
	end_offset := c.readElementOffset(version)
	if end_offset == 0{
		return nil
	}
	prop_count := c.readElementOffset(version)
	prop_length := c.readElementOffset(version)
	id := c.readShortString()

	element := Element{}
	element.id = id

	oldProp := *element.first_property
	
	for i := 0 ; i < prop_count; i++{
		prop := c.readProperty()
		oldProp.next = prop
		oldProp = prop
	}

	if initial_size -c.Buffered() >- end_offset{
		return element
	}
	BLOCK_SENTINEL_LENGTH = 13
	if version >= 7500{
		BLOCK_SENTINEL_LENGTH = 25
	}

	link := &element.child

	for initial_size- c.buffered() < end_offset -BLOCK_SENTINEL_LENGTH {
		child := c.readElement(version)
		if child==nil{
			fmt.Println(errors.NewError("ReadingChild element failed"))
		}
		*link = child
		link = &(*link).sibling
	}
	c.Discard(BLOCK_SENTINEL_LENGTH)
	return element
}

static OptionalError<Property*> readTextProperty(Cursor* cursor) {
	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.value.is_binary = false;
	prop.next = nullptr;
	if (*cursor.current == '"') {
		prop.type = 'S';
		++cursor.current;
		prop.value.begin = cursor.current;
		while (cursor.current < cursor.end && *cursor.current != '"') {
			++cursor.current;
		}
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '"'
		return prop.release();
	}
	
	if (isdigit(*cursor.current) || *cursor.current == '-') {
		prop.type = 'L';
		prop.value.begin = cursor.current;
		if (*cursor.current == '-') ++cursor.current;
		while (cursor.current < cursor.end && isdigit(*cursor.current)) {
			++cursor.current;
		}
		prop.value.end = cursor.current;

		if (cursor.current < cursor.end && *cursor.current == '.') {
			prop.type = 'D';
			++cursor.current;
			while (cursor.current < cursor.end && isdigit(*cursor.current)) {
				++cursor.current;
			}
			if (cursor.current < cursor.end && (*cursor.current == 'e' || *cursor.current == 'E')) {
				// 10.5e-013
				++cursor.current;
				if (cursor.current < cursor.end && *cursor.current == '-') ++cursor.current;
				while (cursor.current < cursor.end && isdigit(*cursor.current)) ++cursor.current;
			}

			prop.value.end = cursor.current;
		}
		return prop.release();
	}
	
	if (*cursor.current == 'T' || *cursor.current == 'Y') {
		// WTF is this
		prop.type = *cursor.current;
		prop.value.begin = cursor.current;
		++cursor.current;
		prop.value.end = cursor.current;
		return prop.release();
	}

	if (*cursor.current == '*') {
		prop.type = 'l';
		++cursor.current;
		// Vertices: *10740 { a: 14.2760353088379,... }
		while (cursor.current < cursor.end && *cursor.current != ':') {
			++cursor.current;
		}
		if (cursor.current < cursor.end) ++cursor.current; // skip ':'
		skipInsignificantWhitespaces(cursor);
		prop.value.begin = cursor.current;
		prop.count = 0;
		bool is_any = false;
		while (cursor.current < cursor.end && *cursor.current != '}') {
			if (*cursor.current == ',') {
				if (is_any) ++prop.count;
				is_any = false;
			}
			else if (!isspace(*cursor.current) && *cursor.current != '\n') is_any = true;
			if (*cursor.current == '.') prop.type = 'd';
			++cursor.current;
		}
		if (is_any) ++prop.count;
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '}'
		return prop.release();
	}

	assert(false);
	return Error("TODO");
}

static OptionalError<Element*> readTextElement(Cursor* cursor) {
	DataView id = readTextToken(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");
	if(*cursor.current != ':') return Error("Unexpected end of file");
	++cursor.current;

	skipWhitespaces(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");

	Element* element = new Element;
	element.id = id;

	Property** prop_link = &element.first_property;
	while (cursor.current < cursor.end && *cursor.current != '\n' && *cursor.current != '{') {
		OptionalError<Property*> prop = readTextProperty(cursor);
		if (prop.isError()) {
			deleteElement(element);
			return Error();
		}
		if (cursor.current < cursor.end && *cursor.current == ',') {
			++cursor.current;
			skipWhitespaces(cursor);
		}
		skipInsignificantWhitespaces(cursor);

		*prop_link = prop.getValue();
		prop_link = &(*prop_link).next;
	}
	
	Element** link = &element.child;
	if (*cursor.current == '{') {
		++cursor.current;
		skipWhitespaces(cursor);
		while (cursor.current < cursor.end && *cursor.current != '}') {
			OptionalError<Element*> child = readTextElement(cursor);
			if (child.isError()) {
				deleteElement(element);
				return Error();
			}
			skipWhitespaces(cursor);

			*link = child.getValue();
			link = &(*link).sibling;
		}
		if (cursor.current < cursor.end) ++cursor.current; // skip '}'
	}
	return element;
}

func tokenizeText(data []byte, size int) (*Element, error) {
	cursor := NewCursor(data[:size])
	root := &Element{}
	element := &root.child
	for cursor.cur < len(cursor.data) {
		v := cursor.data[cursor.cur]
		if (v == ';' || v == '\r' || v == '\n') {
			skipLine(cursor)
		} else {
			child, err := readTextElement(cursor)
			if err != nil {
				deleteElement(root)
				return nil, err
			}
			*element = child.getValue()
			if element == nil {
				return root, nil
			}
			element = element.sibling
		}
	}
	return root, nil
}

func tokenize(data []byte) *Element, errors.Error{
	cursor := NewCursor(data)

	var header Header
	binary.Read(bytes.NewReader(data), binary.BigEndian, &header)
	cursor.cur += HEADER_BYTES
	
	root := &Element{}
	element := &root.child
	for true {
		child, err := readElement(&cursor, header.version)
		if err != nil {
			deleteElement(root)
			return err
		}
		*element = child
		if element == nil{
			return root
		}
		element = element.sibling
	}
}