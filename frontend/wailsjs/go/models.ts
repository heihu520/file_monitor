export namespace main {
	
	export class DirInsight {
	    totalSize: string;
	    totalBytes: number;
	    fileCount: number;
	    dirCount: number;
	    categories: Record<string, number>;
	    extDetails: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new DirInsight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSize = source["totalSize"];
	        this.totalBytes = source["totalBytes"];
	        this.fileCount = source["fileCount"];
	        this.dirCount = source["dirCount"];
	        this.categories = source["categories"];
	        this.extDetails = source["extDetails"];
	    }
	}
	export class DiskInfo {
	    total: string;
	    free: string;
	    used: string;
	    usage: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.free = source["free"];
	        this.used = source["used"];
	        this.usage = source["usage"];
	    }
	}
	export class FileStat {
	    name: string;
	    path: string;
	    size: string;
	    bytes: number;
	    timeDetail: string;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new FileStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.bytes = source["bytes"];
	        this.timeDetail = source["timeDetail"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

