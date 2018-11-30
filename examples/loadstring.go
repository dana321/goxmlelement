package main

import(
	"github.com/dana321/goxmlelement"

)

const(
	testXMLString1=`<root><class name="johnny"><variable name="name" value="Johnny"/></class></root>`
)

func main(){
	// create a new ElementReader object
	er :=goxmlelement.ElementReader{}
	
	// load variable from string
	er.LoadString(testXMLString1)
	
	// dump the nodes out to terminal
	er.Root.WalkDump()
}
