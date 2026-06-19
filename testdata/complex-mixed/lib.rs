use serde::Deserialize;
use tokio::runtime::Runtime;

#[derive(Deserialize)]
struct Config {
    name: String,
}

fn main() {
    let rt = Runtime::new().unwrap();
    println!("Mixed project");
}
