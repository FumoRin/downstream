export namespace downloader {
	
	export class DownloadState {
	    id: string;
	    url: string;
	    filename: string;
	    category: string;
	    totalSize: number;
	    currentSize: number;
	    percentage: number;
	    status: number;
	    // Go type: time
	    scheduledAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new DownloadState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.category = source["category"];
	        this.totalSize = source["totalSize"];
	        this.currentSize = source["currentSize"];
	        this.percentage = source["percentage"];
	        this.status = source["status"];
	        this.scheduledAt = this.convertValues(source["scheduledAt"], null);
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
	export class Settings {
	    DownloadDir: string;
	    MaxConcurrencyDownload: number;
	    PartsPerDownload: number;
	    MinMultiPartDownload: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DownloadDir = source["DownloadDir"];
	        this.MaxConcurrencyDownload = source["MaxConcurrencyDownload"];
	        this.PartsPerDownload = source["PartsPerDownload"];
	        this.MinMultiPartDownload = source["MinMultiPartDownload"];
	    }
	}
	export class TargetInfo {
	    Filename: string;
	    TotalSize: number;
	    SupportMultiPart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TargetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Filename = source["Filename"];
	        this.TotalSize = source["TotalSize"];
	        this.SupportMultiPart = source["SupportMultiPart"];
	    }
	}

}

