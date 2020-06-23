extern crate curl;

use curl::easy::Easy;
use std::env;
use std::io::{self, Write};

fn main() {
    let args: Vec<String> = env::args().collect();

    let key = "N";
    let n = match env::var(key) {
        Ok(val) => val.parse().unwrap(),
        Err(_e) => 1000,
    };

    for _ in 0..n {
	let mut handle = Easy::new();
	handle.ssl_verify_host(false).unwrap();
	handle.ssl_verify_peer(false).unwrap();

        handle.url(&args[1].clone()).unwrap();
        handle
            .write_function(|data| Ok(io::sink().write(data).unwrap()))
            .unwrap();
        handle.perform().unwrap();

        let time_total = handle.total_time().unwrap();
        let time_namelookup = handle.namelookup_time().unwrap();
        let time_connect = handle.connect_time().unwrap();
        let time_pretransfer = handle.pretransfer_time().unwrap();
        let time_starttransfer = handle.starttransfer_time().unwrap();
        let http_code = handle.response_code().unwrap();

        println!(
            "namelookup:{:6} connect:{:6} pretransfer:{:6} starttransfer:{:6} total:{:6} httpcode: {:<4}",
            time_namelookup.as_millis(),
            time_connect.as_millis(),
            time_pretransfer.as_millis(),
            time_starttransfer.as_millis(),
            time_total.as_millis(),
            http_code,
        );
    }
}
