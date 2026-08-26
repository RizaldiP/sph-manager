export namespace collaboration {
	
	export class DiscoveredRoom {
	    roomId: string;
	    roomName: string;
	    documentNumber: string;
	    projectName: string;
	    hostIP: string;
	    hostName: string;
	    port: number;
	    users: number;
	    // Go type: time
	    lastSeen: any;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredRoom(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roomId = source["roomId"];
	        this.roomName = source["roomName"];
	        this.documentNumber = source["documentNumber"];
	        this.projectName = source["projectName"];
	        this.hostIP = source["hostIP"];
	        this.hostName = source["hostName"];
	        this.port = source["port"];
	        this.users = source["users"];
	        this.lastSeen = this.convertValues(source["lastSeen"], null);
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
	export class Participant {
	    id: string;
	    displayName: string;
	    deviceName: string;
	    role: string;
	    // Go type: time
	    joinedAt: any;
	    // Go type: time
	    lastSeen: any;
	
	    static createFrom(source: any = {}) {
	        return new Participant(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.deviceName = source["deviceName"];
	        this.role = source["role"];
	        this.joinedAt = this.convertValues(source["joinedAt"], null);
	        this.lastSeen = this.convertValues(source["lastSeen"], null);
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
	export class RoomInfo {
	    roomId: string;
	    sphDocumentId: number;
	    documentNumber: string;
	    projectName: string;
	    roomCode: string;
	    roomName: string;
	    accessCode?: string;
	    hostName: string;
	    hostDevice: string;
	    hostIPs?: string[];
	    port: number;
	    status: string;
	    version: number;
	    participants?: Participant[];
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RoomInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roomId = source["roomId"];
	        this.sphDocumentId = source["sphDocumentId"];
	        this.documentNumber = source["documentNumber"];
	        this.projectName = source["projectName"];
	        this.roomCode = source["roomCode"];
	        this.roomName = source["roomName"];
	        this.accessCode = source["accessCode"];
	        this.hostName = source["hostName"];
	        this.hostDevice = source["hostDevice"];
	        this.hostIPs = source["hostIPs"];
	        this.port = source["port"];
	        this.status = source["status"];
	        this.version = source["version"];
	        this.participants = this.convertValues(source["participants"], Participant);
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
	export class UISnapshot {
	    mode: string;
	    connection?: string;
	    room?: RoomInfo;
	    doc?: number[];
	    participants?: Participant[];
	    activities?: services.CollabActivity[];
	    version?: number;
	    error?: string;
	    notice?: string;
	
	    static createFrom(source: any = {}) {
	        return new UISnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.connection = source["connection"];
	        this.room = this.convertValues(source["room"], RoomInfo);
	        this.doc = source["doc"];
	        this.participants = this.convertValues(source["participants"], Participant);
	        this.activities = this.convertValues(source["activities"], services.CollabActivity);
	        this.version = source["version"];
	        this.error = source["error"];
	        this.notice = source["notice"];
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

export namespace importers {
	
	export class ColumnMapping {
	    nameCol: number;
	    nameSpan: number;
	    qtyCol: number;
	    unitCol: number;
	    serviceCol: number;
	    materialCol: number;
	    unitPriceCol: number;
	    unitPriceAs?: string;
	    serviceTotal: boolean;
	    materialTotal: boolean;
	    headerRows: number;
	    notes?: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nameCol = source["nameCol"];
	        this.nameSpan = source["nameSpan"];
	        this.qtyCol = source["qtyCol"];
	        this.unitCol = source["unitCol"];
	        this.serviceCol = source["serviceCol"];
	        this.materialCol = source["materialCol"];
	        this.unitPriceCol = source["unitPriceCol"];
	        this.unitPriceAs = source["unitPriceAs"];
	        this.serviceTotal = source["serviceTotal"];
	        this.materialTotal = source["materialTotal"];
	        this.headerRows = source["headerRows"];
	        this.notes = source["notes"];
	    }
	}
	export class PreviewRow {
	    rowIndex: number;
	    suggested: string;
	    level: string;
	    marker: string;
	    name: string;
	    qty: number;
	    unit: string;
	    servicePrice: number;
	    materialPrice: number;
	    raw: string;
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new PreviewRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowIndex = source["rowIndex"];
	        this.suggested = source["suggested"];
	        this.level = source["level"];
	        this.marker = source["marker"];
	        this.name = source["name"];
	        this.qty = source["qty"];
	        this.unit = source["unit"];
	        this.servicePrice = source["servicePrice"];
	        this.materialPrice = source["materialPrice"];
	        this.raw = source["raw"];
	        this.errors = source["errors"];
	    }
	}
	export class SheetPreview {
	    grid: string[][];
	    totalRows: number;
	    totalCols: number;
	    suggestedMapping: ColumnMapping;
	    notes?: string[];
	    mainCount: number;
	    subCount: number;
	    unknownCount: number;
	
	    static createFrom(source: any = {}) {
	        return new SheetPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grid = source["grid"];
	        this.totalRows = source["totalRows"];
	        this.totalCols = source["totalCols"];
	        this.suggestedMapping = this.convertValues(source["suggestedMapping"], ColumnMapping);
	        this.notes = source["notes"];
	        this.mainCount = source["mainCount"];
	        this.subCount = source["subCount"];
	        this.unknownCount = source["unknownCount"];
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
	export class Vessel {
	    id: number;
	    customerId: number;
	    code: string;
	    name: string;
	    vesselNumber: string;
	    vesselType: string;
	    notes: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    customer?: Customer;
	
	    static createFrom(source: any = {}) {
	        return new Vessel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.customerId = source["customerId"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.vesselNumber = source["vesselNumber"];
	        this.vesselType = source["vesselType"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.customer = this.convertValues(source["customer"], Customer);
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
	export class Customer {
	    id: number;
	    code: string;
	    name: string;
	    address: string;
	    phone: string;
	    email: string;
	    picName: string;
	    picPosition: string;
	    notes: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    vessels?: Vessel[];
	
	    static createFrom(source: any = {}) {
	        return new Customer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.picName = source["picName"];
	        this.picPosition = source["picPosition"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.vessels = this.convertValues(source["vessels"], Vessel);
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
	export class Material {
	    id: number;
	    code: string;
	    name: string;
	    description: string;
	    unit: string;
	    defaultPrice: number;
	    supplier: string;
	    notes: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	
	    static createFrom(source: any = {}) {
	        return new Material(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.unit = source["unit"];
	        this.defaultPrice = source["defaultPrice"];
	        this.supplier = source["supplier"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
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
	export class SphRevision {
	    id: number;
	    sphDocumentId: number;
	    fromDocumentId?: number;
	    revisionNumber: number;
	    note: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SphRevision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sphDocumentId = source["sphDocumentId"];
	        this.fromDocumentId = source["fromDocumentId"];
	        this.revisionNumber = source["revisionNumber"];
	        this.note = source["note"];
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
	export class SphSubItem {
	    id: number;
	    sphItemId: number;
	    sequence: number;
	    nameSnapshot: string;
	    descriptionSnapshot: string;
	    quantity: number;
	    unit: string;
	    weight: number;
	    allocatedValue: number;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    serviceTotal: number;
	    materialTotal: number;
	    total: number;
	    notes: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SphSubItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sphItemId = source["sphItemId"];
	        this.sequence = source["sequence"];
	        this.nameSnapshot = source["nameSnapshot"];
	        this.descriptionSnapshot = source["descriptionSnapshot"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.weight = source["weight"];
	        this.allocatedValue = source["allocatedValue"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.serviceTotal = source["serviceTotal"];
	        this.materialTotal = source["materialTotal"];
	        this.total = source["total"];
	        this.notes = source["notes"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class SphItem {
	    id: number;
	    sphDocumentId: number;
	    sequence: number;
	    workItemId?: number;
	    nameSnapshot: string;
	    descriptionSnapshot: string;
	    quantity: number;
	    unit: string;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    serviceTotal: number;
	    materialTotal: number;
	    total: number;
	    pricingMode: string;
	    notes: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    workItem?: WorkItem;
	    subItems?: SphSubItem[];
	
	    static createFrom(source: any = {}) {
	        return new SphItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sphDocumentId = source["sphDocumentId"];
	        this.sequence = source["sequence"];
	        this.workItemId = source["workItemId"];
	        this.nameSnapshot = source["nameSnapshot"];
	        this.descriptionSnapshot = source["descriptionSnapshot"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.serviceTotal = source["serviceTotal"];
	        this.materialTotal = source["materialTotal"];
	        this.total = source["total"];
	        this.pricingMode = source["pricingMode"];
	        this.notes = source["notes"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.workItem = this.convertValues(source["workItem"], WorkItem);
	        this.subItems = this.convertValues(source["subItems"], SphSubItem);
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
	export class SphDocument {
	    id: number;
	    documentNumber: string;
	    revision: number;
	    // Go type: time
	    date: any;
	    customerId: number;
	    vesselId?: number;
	    projectName: string;
	    subject: string;
	    reference: string;
	    location: string;
	    // Go type: time
	    validUntil?: any;
	    picName: string;
	    status: string;
	    subtotalService: number;
	    subtotalMaterial: number;
	    grandTotal: number;
	    terbilang: string;
	    notes: string;
	    // Go type: time
	    finalizedAt?: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    customer?: Customer;
	    vessel?: Vessel;
	    items?: SphItem[];
	    revisions?: SphRevision[];
	
	    static createFrom(source: any = {}) {
	        return new SphDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.documentNumber = source["documentNumber"];
	        this.revision = source["revision"];
	        this.date = this.convertValues(source["date"], null);
	        this.customerId = source["customerId"];
	        this.vesselId = source["vesselId"];
	        this.projectName = source["projectName"];
	        this.subject = source["subject"];
	        this.reference = source["reference"];
	        this.location = source["location"];
	        this.validUntil = this.convertValues(source["validUntil"], null);
	        this.picName = source["picName"];
	        this.status = source["status"];
	        this.subtotalService = source["subtotalService"];
	        this.subtotalMaterial = source["subtotalMaterial"];
	        this.grandTotal = source["grandTotal"];
	        this.terbilang = source["terbilang"];
	        this.notes = source["notes"];
	        this.finalizedAt = this.convertValues(source["finalizedAt"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.customer = this.convertValues(source["customer"], Customer);
	        this.vessel = this.convertValues(source["vessel"], Vessel);
	        this.items = this.convertValues(source["items"], SphItem);
	        this.revisions = this.convertValues(source["revisions"], SphRevision);
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
	
	
	
	export class TemplateItem {
	    id: number;
	    templateId: number;
	    sequence: number;
	    workItemId: number;
	    notes: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    workItem?: WorkItem;
	
	    static createFrom(source: any = {}) {
	        return new TemplateItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.templateId = source["templateId"];
	        this.sequence = source["sequence"];
	        this.workItemId = source["workItemId"];
	        this.notes = source["notes"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class Template {
	    id: number;
	    code: string;
	    name: string;
	    description: string;
	    notes: string;
	    sequence: number;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    deletedAt?: gorm.DeletedAt;
	    items?: TemplateItem[];
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.notes = source["notes"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.deletedAt = this.convertValues(source["deletedAt"], gorm.DeletedAt);
	        this.items = this.convertValues(source["items"], TemplateItem);
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
	export class CollabActivity {
	    actor: string;
	    action: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new CollabActivity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.actor = source["actor"];
	        this.action = source["action"];
	        this.summary = source["summary"];
	    }
	}
	export class CollabDefaults {
	    deviceName: string;
	    port: number;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new CollabDefaults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceName = source["deviceName"];
	        this.port = source["port"];
	        this.displayName = source["displayName"];
	    }
	}
	export class ConfirmRow {
	    rowIndex: number;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfirmRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowIndex = source["rowIndex"];
	        this.level = source["level"];
	    }
	}
	export class VesselView {
	    id: number;
	    customerId: number;
	    code: string;
	    name: string;
	    vesselNumber: string;
	    vesselType: string;
	    notes: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VesselView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.customerId = source["customerId"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.vesselNumber = source["vesselNumber"];
	        this.vesselType = source["vesselType"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	    }
	}
	export class CustomerView {
	    id: number;
	    code: string;
	    name: string;
	    address: string;
	    phone: string;
	    email: string;
	    picName: string;
	    picPosition: string;
	    notes: string;
	    isActive: boolean;
	    vessels: VesselView[];
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomerView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.picName = source["picName"];
	        this.picPosition = source["picPosition"];
	        this.notes = source["notes"];
	        this.isActive = source["isActive"];
	        this.vessels = this.convertValues(source["vessels"], VesselView);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	export class SphDocumentView {
	    id: number;
	    documentNumber: string;
	    revision: number;
	    date: string;
	    customerId: number;
	    customerName: string;
	    vesselId?: number;
	    vesselName: string;
	    projectName: string;
	    subject: string;
	    status: string;
	    itemCount: number;
	    grandTotal: number;
	    terbilang: string;
	    finalizedAt?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SphDocumentView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.documentNumber = source["documentNumber"];
	        this.revision = source["revision"];
	        this.date = source["date"];
	        this.customerId = source["customerId"];
	        this.customerName = source["customerName"];
	        this.vesselId = source["vesselId"];
	        this.vesselName = source["vesselName"];
	        this.projectName = source["projectName"];
	        this.subject = source["subject"];
	        this.status = source["status"];
	        this.itemCount = source["itemCount"];
	        this.grandTotal = source["grandTotal"];
	        this.terbilang = source["terbilang"];
	        this.finalizedAt = source["finalizedAt"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class DashboardStats {
	    totalSph: number;
	    draftCount: number;
	    finalCount: number;
	    acceptedCount: number;
	    monthValue: number;
	    recent: SphDocumentView[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSph = source["totalSph"];
	        this.draftCount = source["draftCount"];
	        this.finalCount = source["finalCount"];
	        this.acceptedCount = source["acceptedCount"];
	        this.monthValue = source["monthValue"];
	        this.recent = this.convertValues(source["recent"], SphDocumentView);
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
	export class DeleteResult {
	    items: number;
	    subs: number;
	
	    static createFrom(source: any = {}) {
	        return new DeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = source["items"];
	        this.subs = source["subs"];
	    }
	}
	export class HeaderPatch {
	    date: string;
	    customerId: number;
	    vesselId?: number;
	    projectName: string;
	    subject: string;
	    reference: string;
	    location: string;
	    validUntil: string;
	    picName: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderPatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.customerId = source["customerId"];
	        this.vesselId = source["vesselId"];
	        this.projectName = source["projectName"];
	        this.subject = source["subject"];
	        this.reference = source["reference"];
	        this.location = source["location"];
	        this.validUntil = source["validUntil"];
	        this.picName = source["picName"];
	        this.notes = source["notes"];
	    }
	}
	export class ImportResult {
	    itemsCreated: number;
	    subsCreated: number;
	    skipped: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemsCreated = source["itemsCreated"];
	        this.subsCreated = source["subsCreated"];
	        this.skipped = source["skipped"];
	    }
	}
	export class ItemFields {
	    workItemId?: number;
	    name: string;
	    description: string;
	    quantity: number;
	    unit: string;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    pricingMode: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemFields(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workItemId = source["workItemId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.pricingMode = source["pricingMode"];
	        this.notes = source["notes"];
	    }
	}
	export class SubItemFields {
	    name: string;
	    description: string;
	    quantity: number;
	    unit: string;
	    weight: number;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new SubItemFields(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.weight = source["weight"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.notes = source["notes"];
	    }
	}
	export class OpPayload {
	    type: string;
	    itemId?: number;
	    subItemId?: number;
	    toIndex?: number;
	    header?: HeaderPatch;
	    item?: ItemFields;
	    subItem?: SubItemFields;
	
	    static createFrom(source: any = {}) {
	        return new OpPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.itemId = source["itemId"];
	        this.subItemId = source["subItemId"];
	        this.toIndex = source["toIndex"];
	        this.header = this.convertValues(source["header"], HeaderPatch);
	        this.item = this.convertValues(source["item"], ItemFields);
	        this.subItem = this.convertValues(source["subItem"], SubItemFields);
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
	export class SettingsInput {
	    companyName: string;
	    companyCity: string;
	    companyAddress: string;
	    sphNumberFormat: string;
	    signerName: string;
	    signerPosition: string;
	    defaultNotes: string;
	    collabPort: number;
	    collabDisplayName: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.companyName = source["companyName"];
	        this.companyCity = source["companyCity"];
	        this.companyAddress = source["companyAddress"];
	        this.sphNumberFormat = source["sphNumberFormat"];
	        this.signerName = source["signerName"];
	        this.signerPosition = source["signerPosition"];
	        this.defaultNotes = source["defaultNotes"];
	        this.collabPort = source["collabPort"];
	        this.collabDisplayName = source["collabDisplayName"];
	    }
	}
	export class SettingsView {
	    companyName: string;
	    companyCity: string;
	    companyAddress: string;
	    logoPath: string;
	    sphNumberFormat: string;
	    signerName: string;
	    signerPosition: string;
	    defaultNotes: string;
	    collabPort: number;
	    collabDisplayName: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.companyName = source["companyName"];
	        this.companyCity = source["companyCity"];
	        this.companyAddress = source["companyAddress"];
	        this.logoPath = source["logoPath"];
	        this.sphNumberFormat = source["sphNumberFormat"];
	        this.signerName = source["signerName"];
	        this.signerPosition = source["signerPosition"];
	        this.defaultNotes = source["defaultNotes"];
	        this.collabPort = source["collabPort"];
	        this.collabDisplayName = source["collabDisplayName"];
	    }
	}
	
	export class SphHeaderInput {
	    date: string;
	    customerId: number;
	    vesselId?: number;
	    projectName: string;
	    subject: string;
	    reference: string;
	    location: string;
	    validUntil: string;
	    picName: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new SphHeaderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.customerId = source["customerId"];
	        this.vesselId = source["vesselId"];
	        this.projectName = source["projectName"];
	        this.subject = source["subject"];
	        this.reference = source["reference"];
	        this.location = source["location"];
	        this.validUntil = source["validUntil"];
	        this.picName = source["picName"];
	        this.notes = source["notes"];
	    }
	}
	export class SphSubItemInput {
	    id?: number;
	    name: string;
	    description: string;
	    quantity: number;
	    unit: string;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    weight: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new SphSubItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.weight = source["weight"];
	        this.notes = source["notes"];
	    }
	}
	export class SphItemInput {
	    id?: number;
	    workItemId?: number;
	    name: string;
	    description: string;
	    quantity: number;
	    unit: string;
	    serviceUnitPrice: number;
	    materialUnitPrice: number;
	    pricingMode: string;
	    notes: string;
	    subItems: SphSubItemInput[];
	
	    static createFrom(source: any = {}) {
	        return new SphItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workItemId = source["workItemId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unit = source["unit"];
	        this.serviceUnitPrice = source["serviceUnitPrice"];
	        this.materialUnitPrice = source["materialUnitPrice"];
	        this.pricingMode = source["pricingMode"];
	        this.notes = source["notes"];
	        this.subItems = this.convertValues(source["subItems"], SphSubItemInput);
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
	export class SphSaveInput {
	    header: SphHeaderInput;
	    items: SphItemInput[];
	
	    static createFrom(source: any = {}) {
	        return new SphSaveInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.header = this.convertValues(source["header"], SphHeaderInput);
	        this.items = this.convertValues(source["items"], SphItemInput);
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
	
	
	export class TemplateItemInput {
	    workItemId: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workItemId = source["workItemId"];
	        this.notes = source["notes"];
	    }
	}
	export class TemplateView {
	    id: number;
	    code: string;
	    name: string;
	    description: string;
	    notes: string;
	    sequence: number;
	    isActive: boolean;
	    itemCount: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.notes = source["notes"];
	        this.sequence = source["sequence"];
	        this.isActive = source["isActive"];
	        this.itemCount = source["itemCount"];
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

