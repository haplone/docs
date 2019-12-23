pub struct Post{
    state: Option<Box<dyn state>>,
    content: String,
}

impl Post {
    pub fn new() -> Post{
        Post{
            state: Some(Box::new(Draft{})),
            content: String::new(),
        }
    }

    pub fn add_text(&mut self,text: &str){
        self.contect.push_str(text);
    }

    pub fn request_review(&mut self){
        if let Some(s) = self.state.take(){
            self.state = Some(s.request_review())
        }
    }

    pub fn approve(&mut self){
        if let Some(s) = self.state.take(){
            self.state = Some(s.approve())
        }
    }

    pub fn content(&self) -> &str {
        self.state.as_ref().unwrap().content(&self)
    }
}

trait State{
    fn request_review(self: Box<Self>) -> Box<dyn State>;
    fn approve(&mut self) -> Box<dyn State>;
    fn content<'a>
}

struct Draft{}

impl State for Draft{
    fn request_review(self: Box<Self>) -> Box<dyn State>{
        Box::new(PendingReview{})
    }

    fn approve(self: Box<Self>) -> Box<dyn State>{
        self
    }
}

struct PendingReview{}

impl State for PendingReview{
    fn request_review(self: Box<Self>) -> Box<dyn State>{
        self
    }

    fn approve(self: Box<Self>) -> Box<dyn state>{
        Box::new(PUblished{})
    }
}

struct Published{}

impl State for Published{
    fn request_review(self: Box<Self>) -> Box<dyn State>{
        self
    }

    fn approve(self: Box<Self>) -> Box<dyn State>{
        self
    }
}


