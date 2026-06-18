use serde::Deserialize;
use tokio::runtime::Runtime;
use reqwest::blocking::Client;

fn main() {
    let rt = Runtime::new().unwrap();
    let client = Client::new();
    
    #[derive(Deserialize)]
    struct Response {
        message: String,
    }
    
    println!("Rust project example");
}
