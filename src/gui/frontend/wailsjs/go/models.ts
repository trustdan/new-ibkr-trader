export namespace main {
	
	export class FilterConfig {
	    type: string;
	    params: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FilterConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.params = source["params"];
	    }
	}
	export class Option {
	    symbol: string;
	    strike: number;
	    expiry: string;
	    right: string;
	    delta: number;
	    gamma: number;
	    theta: number;
	    vega: number;
	    iv: number;
	    volume: number;
	    open_interest: number;
	    bid_price: number;
	    ask_price: number;
	    last_price: number;
	
	    static createFrom(source: any = {}) {
	        return new Option(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.symbol = source["symbol"];
	        this.strike = source["strike"];
	        this.expiry = source["expiry"];
	        this.right = source["right"];
	        this.delta = source["delta"];
	        this.gamma = source["gamma"];
	        this.theta = source["theta"];
	        this.vega = source["vega"];
	        this.iv = source["iv"];
	        this.volume = source["volume"];
	        this.open_interest = source["open_interest"];
	        this.bid_price = source["bid_price"];
	        this.ask_price = source["ask_price"];
	        this.last_price = source["last_price"];
	    }
	}
	export class ScanRequest {
	    symbol: string;
	    filters: FilterConfig[];
	    max_results: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.symbol = source["symbol"];
	        this.filters = this.convertValues(source["filters"], FilterConfig);
	        this.max_results = source["max_results"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScanResponse {
	    symbol: string;
	    // Go type: time
	    scan_time: any;
	    result_count: number;
	    options: Option[];
	
	    static createFrom(source: any = {}) {
	        return new ScanResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.symbol = source["symbol"];
	        this.scan_time = this.convertValues(source["scan_time"], null);
	        this.result_count = source["result_count"];
	        this.options = this.convertValues(source["options"], Option);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SystemHealth {
	    // Go type: struct { Connected bool "json:\"connected\""; Uptime int64 "json:\"uptime\"" }
	    tws: any;
	    // Go type: struct { Active int "json:\"active\""; Max int "json:\"max\""; UsagePct int "json:\"usage_pct\"" }
	    subscriptions: any;
	    // Go type: struct { Size int "json:\"size\""; Processing bool "json:\"processing\"" }
	    queue: any;
	    throttling: boolean;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new SystemHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tws = this.convertValues(source["tws"], Object);
	        this.subscriptions = this.convertValues(source["subscriptions"], Object);
	        this.queue = this.convertValues(source["queue"], Object);
	        this.throttling = source["throttling"];
	        this.errors = source["errors"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

