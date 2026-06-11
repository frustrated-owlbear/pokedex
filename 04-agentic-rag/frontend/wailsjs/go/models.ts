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

