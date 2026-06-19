use serde::{Deserialize, Serialize};
use serde_json::Value;
use log::info;

mod handlers;

#[derive(Serialize, Deserialize)]
struct Config {
    host: String,
    port: u16,
}

fn main() {
    env_logger::init();
    info!("Starting server");
    
    let config = Config {
        host: "localhost".to_string(),
        port: 8080,
    };
    
    let json: Value = serde_json::to_value(&config).unwrap();
    println!("{}", json);
}
