export namespace domain {
	
	export class CurrentSituation {
	    summary: string;
	    advice: string[];
	
	    static createFrom(source: any = {}) {
	        return new CurrentSituation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.advice = source["advice"];
	    }
	}
	export class KnowledgeArticle {
	    id: string;
	    title: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new KnowledgeArticle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	    }
	}
	export class Pokemon {
	    name: string;
	    id: number;
	    imageUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Pokemon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.imageUrl = source["imageUrl"];
	    }
	}
	export class PokemonDetail {
	    id: number;
	    name: string;
	    imageUrl: string;
	    types: string;
	    region: string;
	    level: number;
	    hp: number;
	    maxHp: number;
	    ability: string;
	    typesList: string[];
	    strengths: string[];
	    weaknesses: string[];
	    resistances: string[];
	
	    static createFrom(source: any = {}) {
	        return new PokemonDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.imageUrl = source["imageUrl"];
	        this.types = source["types"];
	        this.region = source["region"];
	        this.level = source["level"];
	        this.hp = source["hp"];
	        this.maxHp = source["maxHp"];
	        this.ability = source["ability"];
	        this.typesList = source["typesList"];
	        this.strengths = source["strengths"];
	        this.weaknesses = source["weaknesses"];
	        this.resistances = source["resistances"];
	    }
	}
	export class PokemonInfo {
	    id: number;
	    name: string;
	    types: string;
	    region: string;
	
	    static createFrom(source: any = {}) {
	        return new PokemonInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.types = source["types"];
	        this.region = source["region"];
	    }
	}
	export class TrainerProfile {
	    trainerId: string;
	    avatarUrl: string;
	    connectionStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new TrainerProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trainerId = source["trainerId"];
	        this.avatarUrl = source["avatarUrl"];
	        this.connectionStatus = source["connectionStatus"];
	    }
	}
	export class TypeAnalysis {
	    pokemonName: string;
	    types: string[];
	    strengths: string[];
	    weaknesses: string[];
	    resistances: string[];
	
	    static createFrom(source: any = {}) {
	        return new TypeAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pokemonName = source["pokemonName"];
	        this.types = source["types"];
	        this.strengths = source["strengths"];
	        this.weaknesses = source["weaknesses"];
	        this.resistances = source["resistances"];
	    }
	}

}

