use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Serialize, Deserialize)]
pub struct Config {
    pub name: String,
    pub value: i32,
}

pub fn parse_config(json: &str) -> Config {
    serde_json::from_str(json).unwrap()
}
