export namespace downloader {
	
	export class DownloadState {
	    ID: string;
	    URL: string;
	    Filename: string;
	    Category: string;
	    TotalSize: number;
	    Status: number;
	    // Go type: time
	    ScheduledAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new DownloadState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.URL = source["URL"];
	        this.Filename = source["Filename"];
	        this.Category = source["Category"];
	        this.TotalSize = source["TotalSize"];
	        this.Status = source["Status"];
	        this.ScheduledAt = this.convertValues(source["ScheduledAt"], null);
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

}

