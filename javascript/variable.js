var num=100;
function print(){
	console.log(num);
	var num;
}

print();


(function(n1){
	console.log(n1);
	var n1=10;
}(100));


(function(n2){
	console.log(n2);
	var n2=10;
	function n2(){}
}(100));
