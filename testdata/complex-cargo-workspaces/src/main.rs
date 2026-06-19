use core::Config;

fn main() {
    let config = parse_config(r#"{"name": "test", "value": 42}"#);
    println!("Config: {:?}", config);
}
