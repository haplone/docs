var num=100;

var obj={
	num:200,
	inner:{
		num: 300,
		print: function(){
			console.log(this.num);
		}
	}
}

obj.inner.print();

var func = obj.inner.print;
func();

(obj.inner.print)();

(obj.inner.print=obj.inner.print)();
