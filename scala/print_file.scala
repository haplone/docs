import scala.io.Source

if(args.length>0){
	for(line<-Source.fromFile(args(0)).getLines()){
		println(line.length+" \t"+ line)
	}
} else {
	Console.err.println("Please enter a filename")
}
