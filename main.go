package goxmlelement

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"
)


func check(err error) {
    if err != nil {
   	 fmt.Printf("failed with '%s'\n", err)
        panic(err)
    }
}

type Element struct {
	Name string
	Value string
	Attr map[string]string
	Parent *Element
	Children []*Element
}

func (e *Element) AddChild(Name string,Value string,Attr map[string]string) *Element{
	var mpt []*Element
	n := Element{Name,Value,Attr,e,mpt}
	e.Children=append(e.Children,&n)
	return &n
}
func (e *Element) GetPath() string{
	str := ""
	ele:=e
	
	for ele != nil {
		iter:=1
		
		if ele.Parent != nil{
			for k,v := range ele.Parent.Children {
				if v==e {
					iter=k+1
					break
				}
			}
		}
		str="/"+ele.Name+"["+strconv.Itoa(iter)+"]"+str
		ele=ele.Parent
	}
	
	return str
}
func (e *Element) GetChildrenByTagName(tagName string) []*Element{
	var ret []*Element
	 
	tagName=strings.ToLower(tagName)
	 
	for _,el := range e.Children{
		if strings.ToLower(el.Name) == tagName{
			ret=append(ret,el)
		}
	}
	return ret
}
func (e *Element) WalkDump(){
	fmt.Println(e.GetPath(),"=",e.Value,"==",e.Attr)
	
	if len(e.Children)>0{
		for _,x := range e.Children {
			x.WalkDump()
		}
	}	
}
func (e *Element) Walk(each func(*Element)bool){
	var doChildren bool=each(e)
	
	if doChildren && len(e.Children)>0{
		for _,x := range e.Children {
			x.Walk(each)
		}
	}	
}
type ElementReader struct {
	Root *Element
	cv []*Element
}
func (er *ElementReader) LoadFile(fileName string){
	file, err := os.Open(fileName)
	check(err)
	defer file.Close()
	er.LoadStream(file)
}

func (er *ElementReader) LoadString(data string){
	datareader:=strings.NewReader(data)
	er.LoadStream(datareader)
}
func (er *ElementReader) LoadStream(source io.Reader){
	decoder := xml.NewDecoder(source)
	
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			// end of file
			
			break
		}
		check(err)

		switch v := t.(type) {
			case xml.StartElement:

				Attributes:= make(map[string]string)
				for _, Attr := range v.Attr {
					Attributes[Attr.Name.Local]=Attr.Value
				}

				if len(er.cv)==0 {
					var mpt []*Element
					var mp *Element
					newRoot := Element{v.Name.Local,"",Attributes,mp,mpt}
					er.Root=&newRoot
					er.cv=append(er.cv,er.Root)
				} else {
					el :=er.cv[len(er.cv)-1].AddChild(v.Name.Local,"",Attributes)
					er.cv=append(er.cv,el)
				}

			case xml.EndElement:
				if len(er.cv) > 0 {
					er.cv = er.cv[:len(er.cv)-1]
				}
			case xml.CharData:
				cda:=strings.TrimSpace(string(v))
				if cda!="" {
					Attributes:= make(map[string]string)
					er.cv[len(er.cv)-1].AddChild("#text",cda,Attributes)
				}
			case xml.Comment:
				er.cv[len(er.cv)-1].AddChild("#comment",string(v),make(map[string]string))

			case xml.ProcInst:
				if strings.ToLower(string(v.Target))!="xml" && len(er.cv) != 0 {
					// handle XML processing instruction like <?target inst?>
					er.cv[len(er.cv)-1].AddChild("#"+string(v.Target),string(v.Inst),make(map[string]string))
				}
			case xml.Directive:
				// unhandled for now
				//fmt.Printf("Directive: %s\n", string(v))
		}
	}
}
