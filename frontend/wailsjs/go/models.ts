export namespace adapters {
	
	export class MIDIPortInfo {
	    Index: number;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new MIDIPortInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Name = source["Name"];
	    }
	}

}

