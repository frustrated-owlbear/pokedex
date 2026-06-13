export namespace agent {
	
	export class TraceStep {
	    id: string;
	    kind: string;
	    timestamp: string;
	    title: string;
	    detail?: string;
	    tools?: string[];
	
	    static createFrom(source: any = {}) {
	        return new TraceStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.timestamp = source["timestamp"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.tools = source["tools"];
	    }
	}

}

export namespace domain {
	
	export class TeamPokemon {
	    id: number;
	    dexId: number;
	    name: string;
	    level: number;
	    primaryType: string;
	    hp: number;
	    maxHp: number;
	    caughtDate: string;
	    birthday?: string;
	    imageUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new TeamPokemon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.dexId = source["dexId"];
	        this.name = source["name"];
	        this.level = source["level"];
	        this.primaryType = source["primaryType"];
	        this.hp = source["hp"];
	        this.maxHp = source["maxHp"];
	        this.caughtDate = source["caughtDate"];
	        this.birthday = source["birthday"];
	        this.imageUrl = source["imageUrl"];
	    }
	}

}

export namespace main {
	
	export class SituationView {
	    location: string;
	    region: string;
	    time: string;
	    period: string;
	    weather: string;
	    memory: string;
	    tools: string[];
	    analysis?: string;
	
	    static createFrom(source: any = {}) {
	        return new SituationView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.location = source["location"];
	        this.region = source["region"];
	        this.time = source["time"];
	        this.period = source["period"];
	        this.weather = source["weather"];
	        this.memory = source["memory"];
	        this.tools = source["tools"];
	        this.analysis = source["analysis"];
	    }
	}

}

