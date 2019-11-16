pub trait Draw{
    fn draw(&self);
}

pub struct Screen{
    pub components: Vec<Box<dyn Draw>>,
}

impl Screen{
    pub fn run(&self){
        for c in self.components.iter(){
            c.draw();
        }
    }
}

pub struct Button{
    pub width: u32,
    pub height: u32,
    pub label: String,
}

impl Draw for Button{
    fn draw(&self) {
        println!("Button: w:{},height:{},label:{}",self.width,self.height,self.label);
    }
}

pub struct SelectBox{
    width: u32,
    height: u32,
    options: Vec<String>,
}

impl Draw for SelectBox{
    fn draw(&self){
        println!("SelectBox: w:{},h:{},o:{:?}",self.width,self.height,self.options);
    }
}

fn main(){
    let screen = Screen{
        components: vec![
            Box::new(SelectBox{
                width:75,
                height: 10,
                options: vec![
                    String::from("Yes"),
                    String::from("Maybe"),
                    String::from("No"),
                ],
            }),
            Box::new(Button{
                width: 50,
                height: 10,
                label:String::from("Ok"),
            }),
        ],
    };

    screen.run();
}
