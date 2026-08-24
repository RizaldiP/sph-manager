export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

export namespace models {
	
	export class WorkSubItem {
	    id: number;
	    workItemId: number;
	    code: string;
	    sequence: number;
	    name: string;
	    description: string;
	    difficultyWeight: number;
	    defaultUnit: string;
	    defaultQuantity: number;
	    defaultServicePrice: number;
	    defaultMaterialPrice: number;
	    notes: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    workItem?: WorkItem;
	
	    static createFrom(source: any = {}) {
	        return new WorkSubItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workItemId = source["workItemId"];
	        this.code = source["code"];
	        this.sequence = source["sequence"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.difficultyWeight = source["difficultyWeight"];
	        this.defaultUnit = source["defaultUnit"];
	        this.defaultQuantity = source["defaultQuantity"];
	        this.defaultServicePrice = source["defaultServicePrice"];
	        this.defaultMaterialPrice = source["defaultMaterialPrice"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.workItem = this.convertValues(source["workItem"], WorkItem);
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
	export class WorkItem {
	    id: number;
	    categoryId: number;
	    code: string;
	    name: string;
	    description: string;
	    defaultUnit: string;
	    defaultQuantity: number;
	    defaultServicePrice: number;
	    defaultMaterialPrice: number;
	    notes: string;
	    sequence: number;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    category?: Category;
	    subItems?: WorkSubItem[];
	
	    static createFrom(source: any = {}) {
	        return new WorkItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.categoryId = source["categoryId"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.defaultUnit = source["defaultUnit"];
	        this.defaultQuantity = source["defaultQuantity"];
	        this.defaultServicePrice = source["defaultServicePrice"];
	        this.defaultMaterialPrice = source["defaultMaterialPrice"];
	        this.notes = source["notes"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.category = this.convertValues(source["category"], Category);
	        this.subItems = this.convertValues(source["subItems"], WorkSubItem);
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
	export class Category {
	    id: number;
	    code: string;
	    name: string;
	    description: string;
	    sequence: number;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    workItems?: WorkItem[];
	
	    static createFrom(source: any = {}) {
	        return new Category(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.workItems = this.convertValues(source["workItems"], WorkItem);
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

export namespace services {
	
	export class CategoryView {
	    id: number;
	    code: string;
	    name: string;
	    description: string;
	    sequence: number;
	    isActive: boolean;
	    workItemCount: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CategoryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.workItemCount = source["workItemCount"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class WorkItemView {
	    id: number;
	    categoryId: number;
	    code: string;
	    name: string;
	    description: string;
	    defaultUnit: string;
	    defaultQuantity: number;
	    defaultServicePrice: number;
	    defaultMaterialPrice: number;
	    notes: string;
	    sequence: number;
	    isActive: boolean;
	    subItemCount: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkItemView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.categoryId = source["categoryId"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.defaultUnit = source["defaultUnit"];
	        this.defaultQuantity = source["defaultQuantity"];
	        this.defaultServicePrice = source["defaultServicePrice"];
	        this.defaultMaterialPrice = source["defaultMaterialPrice"];
	        this.notes = source["notes"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.subItemCount = source["subItemCount"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

