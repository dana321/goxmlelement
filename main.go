package goxmlelement

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"
)

var GetVarReplFunc func(*Element,string)string

func init(){
	GetVarReplFunc=func(e *Element,s string)string{
		return s
	}
}

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
	Var map[string]interface{}
}
func (e *Element) HasAttrs() bool{
	return len(e.Attr)!=0
}
func (e *Element) HasChildren() bool{
	return len(e.Children)!=0
}
func (e *Element) SetAttr(Name string,Value string){
	e.Attr[Name]=Value
}
func (e *Element) GetAttr(Name string) (string) {
	if val, ok := e.Attr[Name]; ok {
		return GetVarReplFunc(e,val)
	}
	return ""
}
func (e *Element) GetAttrInt(Name string) (int) {
	strval:=GetVarReplFunc(e,e.GetAttr(Name))
	
	i, err := strconv.Atoi(strval)
    
    if err != nil {
    	i=0
    }
    return i
}

func (e *Element) GetVar(Name string) (interface{}) {
	ele:=e
	
	for ele != nil {
		if val, ok := ele.Var[Name]; ok {
			return val
		}
		ele=ele.Parent
	}
	return nil
}
func (e *Element) SetVar(Name string,Value interface{}){
	e.SetVarScope(Name,Value,-1)
}
func (e *Element) SetVarScope(Name string,Value interface{},scope int){
	if scope==0{
		e.Var[Name]=Value
	}else{
		ele:=e
		
		for i := 0; i < -scope; i++ {
			if ele.Parent != nil{
				ele=ele.Parent
			}
		}
		ele.Var[Name]=Value
	}
}
func (e *Element) AddChild(Name string,Value string,Attr map[string]string) *Element{
	var mpt []*Element
	mptvar :=make(map[string]interface{})

	n := Element{Name:Name,Value:Value,Attr:Attr,Parent:e,Children:mpt,Var:mptvar}
	e.Children=append(e.Children,&n)
	return &n
}
func (e *Element) GetPath() string{
	str := ""
	ele:=e
	
	for ele != nil {
		iter:=0
		
		if ele.Parent != nil{
			for _,v := range ele.Parent.Children {
				if v.Name==ele.Name {
					iter++
					if ele==v{
						break
					}
				}
			}
		}else{
			iter++
		}
		str="/"+ele.Name+"["+strconv.Itoa(iter)+"]"+str
		ele=ele.Parent
	}
	
	return str
}
func (e *Element) GetChildByTagName(TagName string) *Element{
	mpt:=new(Element)
	mpt.Name="#empty"
	 
	TagName=strings.ToLower(TagName)
	 
	for _,el := range e.Children{
		if strings.ToLower(el.Name) == TagName{
			return el
		}
	}
	return mpt
}
func (e *Element) GetChildrenByTagName(TagName string) []*Element{
	var ret []*Element
	 
	TagName=strings.ToLower(TagName)
	 
	for _,el := range e.Children{
		if strings.ToLower(el.Name) == TagName{
			ret=append(ret,el)
		}
	}
	return ret
}
func (e *Element) WalkDump(){
	fmt.Println(e.GetPath()+"="+e.Value)
	if len(e.Attr)>0{
		for k,v:=range e.Attr{
			if len(v)>30{
				v=v[:30]+"...(truncated)"
			}
			fmt.Println("@"+k+"="+v)
		}
	}
	if len(e.Var)>0{
		for k,val:=range e.Var{
			v:=fmt.Sprintf("%v", val)
			
			if len(v)>30{
				v=v[:30]+"...(truncated)"
			}
			fmt.Println("$"+k+"="+v)
		}
	}
	
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
					AttrName:=Attr.Name.Local
					if Attr.Name.Space!=""{
						AttrName=Attr.Name.Space+":"+AttrName
					}
					Attributes[AttrName]=Attr.Value
				}
				TName:=v.Name.Local

				if v.Name.Space!=""{
					TName=v.Name.Space+":"+TName
				}
				
				if len(er.cv)==0 {
					var mpt []*Element
					var mp *Element
					mptvar :=make(map[string]interface{})
					
					newRoot := Element{Name:TName,Value:"",Attr:Attributes,Parent:mp,Children:mpt,Var:mptvar}
					er.Root=&newRoot
					er.cv=append(er.cv,er.Root)
				} else {
					el :=er.cv[len(er.cv)-1].AddChild(TName,"",Attributes)
					er.cv=append(er.cv,el)
				}
			case xml.EndElement:
				if len(er.cv) > 0 {
					er.cv = er.cv[:len(er.cv)-1]
				}
			case xml.CharData:
				cda:=string(v)
				cda=strings.TrimFunc(cda, func(r rune) bool {
					retval:=false
					
					if r==rune('\n') || r==rune('\t') ||  r==rune('\r'){
						retval=true
					}
					return retval
				})
				
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
