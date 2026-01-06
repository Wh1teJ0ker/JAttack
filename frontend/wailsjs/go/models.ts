export namespace db {
	
	export class Asset {
	    id: number;
	    ip: string;
	    os: string;
	    alive: boolean;
	    // Go type: time
	    last_scan_time: any;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ip = source["ip"];
	        this.os = source["os"];
	        this.alive = source["alive"];
	        this.last_scan_time = this.convertValues(source["last_scan_time"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class AssetPort {
	    id: number;
	    asset_id: number;
	    port: number;
	    protocol: string;
	    service: string;
	    product: string;
	    version: string;
	    banner: string;
	    state: string;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AssetPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.asset_id = source["asset_id"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.service = source["service"];
	        this.product = source["product"];
	        this.version = source["version"];
	        this.banner = source["banner"];
	        this.state = source["state"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class AuthResult {
	    id: number;
	    asset_id: number;
	    port_id: number;
	    service_type: string;
	    username: string;
	    password: string;
	    success: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AuthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.asset_id = source["asset_id"];
	        this.port_id = source["port_id"];
	        this.service_type = source["service_type"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.success = source["success"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class SensitiveResult {
	    id: number;
	    web_service_id: number;
	    source_file: string;
	    info_type: string;
	    content: string;
	    context: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new SensitiveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.web_service_id = source["web_service_id"];
	        this.source_file = source["source_file"];
	        this.info_type = source["info_type"];
	        this.content = source["content"];
	        this.context = source["context"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class WebDirectory {
	    id: number;
	    web_service_id: number;
	    path: string;
	    status_code: number;
	    content_length: number;
	    title: string;
	    content_type: string;
	    redirect_url: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WebDirectory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.web_service_id = source["web_service_id"];
	        this.path = source["path"];
	        this.status_code = source["status_code"];
	        this.content_length = source["content_length"];
	        this.title = source["title"];
	        this.content_type = source["content_type"];
	        this.redirect_url = source["redirect_url"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class WebJSFile {
	    id: number;
	    web_service_id: number;
	    path: string;
	    full_url: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WebJSFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.web_service_id = source["web_service_id"];
	        this.path = source["path"];
	        this.full_url = source["full_url"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class WebService {
	    id: number;
	    asset_id: number;
	    port_id: number;
	    url: string;
	    title: string;
	    server: string;
	    fingerprints: string;
	    screenshot_path: string;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WebService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.asset_id = source["asset_id"];
	        this.port_id = source["port_id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.server = source["server"];
	        this.fingerprints = source["fingerprints"];
	        this.screenshot_path = source["screenshot_path"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

export namespace infogather {
	
	export class BruteForceTarget {
	    ip: string;
	    port: number;
	    protocol: string;
	    service: string;
	
	    static createFrom(source: any = {}) {
	        return new BruteForceTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.service = source["service"];
	    }
	}
	export class BruteForceConfig {
	    user_dict: string;
	    pass_dict: string;
	    threads: number;
	    timeout: number;
	    protocols: string[];
	    targets: BruteForceTarget[];
	
	    static createFrom(source: any = {}) {
	        return new BruteForceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_dict = source["user_dict"];
	        this.pass_dict = source["pass_dict"];
	        this.threads = source["threads"];
	        this.timeout = source["timeout"];
	        this.protocols = source["protocols"];
	        this.targets = this.convertValues(source["targets"], BruteForceTarget);
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
	
	export class Dictionary {
	    name: string;
	    path: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new Dictionary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	    }
	}
	export class DirScanConfig {
	    target: string;
	    extensions: string[];
	    threads: number;
	    timeout: number;
	    exclude_404: boolean;
	    redirects: boolean;
	    custom_dict: string;
	    recursion_depth: number;
	
	    static createFrom(source: any = {}) {
	        return new DirScanConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.extensions = source["extensions"];
	        this.threads = source["threads"];
	        this.timeout = source["timeout"];
	        this.exclude_404 = source["exclude_404"];
	        this.redirects = source["redirects"];
	        this.custom_dict = source["custom_dict"];
	        this.recursion_depth = source["recursion_depth"];
	    }
	}
	export class DirScanResult {
	    url: string;
	    status: number;
	    size: number;
	    location?: string;
	    fingerprint?: string;
	    title?: string;
	    content_type?: string;
	
	    static createFrom(source: any = {}) {
	        return new DirScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.status = source["status"];
	        this.size = source["size"];
	        this.location = source["location"];
	        this.fingerprint = source["fingerprint"];
	        this.title = source["title"];
	        this.content_type = source["content_type"];
	    }
	}
	export class InfoGathering {
	    id: number;
	    target: string;
	    info_type: string;
	    content: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new InfoGathering(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.target = source["target"];
	        this.info_type = source["info_type"];
	        this.content = source["content"];
	        this.created_at = source["created_at"];
	    }
	}
	export class JSFindResult {
	    url: string;
	    endpoints: string[];
	    js_files: string[];
	    sensitive_info: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new JSFindResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.endpoints = source["endpoints"];
	        this.js_files = source["js_files"];
	        this.sensitive_info = source["sensitive_info"];
	        this.error = source["error"];
	    }
	}
	export class JSFinderOptions {
	    deep_scan: boolean;
	    active_scan: boolean;
	    danger_filter: boolean;
	    concurrency: number;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new JSFinderOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deep_scan = source["deep_scan"];
	        this.active_scan = source["active_scan"];
	        this.danger_filter = source["danger_filter"];
	        this.concurrency = source["concurrency"];
	        this.timeout = source["timeout"];
	    }
	}
	export class ScanConfig {
	    target: string;
	    ports: string;
	    concurrency: number;
	    timeout: number;
	    skip_alive_check: boolean;
	    enable_icmp: boolean;
	    enable_ping: boolean;
	    enable_udp: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.ports = source["ports"];
	        this.concurrency = source["concurrency"];
	        this.timeout = source["timeout"];
	        this.skip_alive_check = source["skip_alive_check"];
	        this.enable_icmp = source["enable_icmp"];
	        this.enable_ping = source["enable_ping"];
	        this.enable_udp = source["enable_udp"];
	    }
	}

}

export namespace logs {
	
	export class LogFile {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new LogFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}

}

export namespace vuln {
	
	export class Vulnerability {
	    id: number;
	    name: string;
	    product: string;
	    vuln_type: string;
	    severity: string;
	    description: string;
	    details: string;
	    status: string;
	    poc_type: string;
	    poc_content: string;
	    reference: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Vulnerability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.product = source["product"];
	        this.vuln_type = source["vuln_type"];
	        this.severity = source["severity"];
	        this.description = source["description"];
	        this.details = source["details"];
	        this.status = source["status"];
	        this.poc_type = source["poc_type"];
	        this.poc_content = source["poc_content"];
	        this.reference = source["reference"];
	        this.created_at = source["created_at"];
	    }
	}

}

