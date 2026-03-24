export namespace database {
	
	export class BudgetStatus {
	    monthlyBudgetGb: number;
	    usedGb: number;
	    usedPercent: number;
	    remainingGb: number;
	    costPerGb: number;
	    projectedGb: number;
	
	    static createFrom(source: any = {}) {
	        return new BudgetStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.monthlyBudgetGb = source["monthlyBudgetGb"];
	        this.usedGb = source["usedGb"];
	        this.usedPercent = source["usedPercent"];
	        this.remainingGb = source["remainingGb"];
	        this.costPerGb = source["costPerGb"];
	        this.projectedGb = source["projectedGb"];
	    }
	}
	export class CacheStats {
	    memoryUsedMb: number;
	    diskUsedMb: number;
	    entries: number;
	    hitCount: number;
	    missCount: number;
	    hitRatio: number;
	    bytesSaved: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memoryUsedMb = source["memoryUsedMb"];
	        this.diskUsedMb = source["diskUsedMb"];
	        this.entries = source["entries"];
	        this.hitCount = source["hitCount"];
	        this.missCount = source["missCount"];
	        this.hitRatio = source["hitRatio"];
	        this.bytesSaved = source["bytesSaved"];
	    }
	}
	export class CostSummary {
	    costToday: number;
	    costWeek: number;
	    costMonth: number;
	    costTotal: number;
	    savedBytes: number;
	    savedCost: number;
	
	    static createFrom(source: any = {}) {
	        return new CostSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.costToday = source["costToday"];
	        this.costWeek = source["costWeek"];
	        this.costMonth = source["costMonth"];
	        this.costTotal = source["costTotal"];
	        this.savedBytes = source["savedBytes"];
	        this.savedCost = source["savedCost"];
	    }
	}
	export class OutputProxy {
	    proxyId: number;
	    localAddr: string;
	    localPort: number;
	    upstream: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new OutputProxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.localAddr = source["localAddr"];
	        this.localPort = source["localPort"];
	        this.upstream = source["upstream"];
	        this.type = source["type"];
	    }
	}
	export class Proxy {
	    id: number;
	    address: string;
	    username: string;
	    password: string;
	    type: string;
	    category: string;
	    enabled: boolean;
	    weight: number;
	    totalBytesUp: number;
	    totalBytesDown: number;
	    totalRequests: number;
	    failCount: number;
	    avgLatencyMs: number;
	    // Go type: time
	    lastCheckAt?: any;
	    lastError: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Proxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.type = source["type"];
	        this.category = source["category"];
	        this.enabled = source["enabled"];
	        this.weight = source["weight"];
	        this.totalBytesUp = source["totalBytesUp"];
	        this.totalBytesDown = source["totalBytesDown"];
	        this.totalRequests = source["totalRequests"];
	        this.failCount = source["failCount"];
	        this.avgLatencyMs = source["avgLatencyMs"];
	        this.lastCheckAt = this.convertValues(source["lastCheckAt"], null);
	        this.lastError = source["lastError"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class ProxyStatus {
	    running: boolean;
	    httpPort: number;
	    socks5Port: number;
	    uptime: number;
	    connections: number;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.httpPort = source["httpPort"];
	        this.socks5Port = source["socks5Port"];
	        this.uptime = source["uptime"];
	        this.connections = source["connections"];
	    }
	}
	export class RealtimeStats {
	    bytesPerSecond: number;
	    residentialBps: number;
	    totalToday: number;
	    residentialToday: number;
	    costToday: number;
	    cacheHitRatio: number;
	    activeConnections: number;
	
	    static createFrom(source: any = {}) {
	        return new RealtimeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bytesPerSecond = source["bytesPerSecond"];
	        this.residentialBps = source["residentialBps"];
	        this.totalToday = source["totalToday"];
	        this.residentialToday = source["residentialToday"];
	        this.costToday = source["costToday"];
	        this.cacheHitRatio = source["cacheHitRatio"];
	        this.activeConnections = source["activeConnections"];
	    }
	}
	export class Rule {
	    id: number;
	    ruleType: string;
	    pattern: string;
	    action: string;
	    priority: number;
	    enabled: boolean;
	    hitCount: number;
	    bytesSaved: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ruleType = source["ruleType"];
	        this.pattern = source["pattern"];
	        this.action = source["action"];
	        this.priority = source["priority"];
	        this.enabled = source["enabled"];
	        this.hitCount = source["hitCount"];
	        this.bytesSaved = source["bytesSaved"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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

