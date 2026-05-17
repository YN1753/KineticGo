export namespace model {
	
	export class Task {
	    ID: number;
	    Name: string;
	    Type: string;
	    Description: string;
	    ExecMode: string;
	    Config: string;
	    // Go type: time
	    CreatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Description = source["Description"];
	        this.ExecMode = source["ExecMode"];
	        this.Config = source["Config"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
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
	export class TaskSchedule {
	    ID: number;
	    Name: string;
	    TaskType: string;
	    CronExpr: string;
	    IntervalSecs: number;
	    IsEnabled: boolean;
	    // Go type: time
	    NextRunTime: any;
	    Config: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskSchedule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.TaskType = source["TaskType"];
	        this.CronExpr = source["CronExpr"];
	        this.IntervalSecs = source["IntervalSecs"];
	        this.IsEnabled = source["IsEnabled"];
	        this.NextRunTime = this.convertValues(source["NextRunTime"], null);
	        this.Config = source["Config"];
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

