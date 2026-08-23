export namespace main {
	
	export class HealthInfo {
	    status: string;
	    version: string;
	    platform: string;
	    databasePath: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.version = source["version"];
	        this.platform = source["platform"];
	        this.databasePath = source["databasePath"];
	    }
	}

}

