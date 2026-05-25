export namespace model {
	
	export class Book {
	    title: string;
	    author: string;
	    cover: string;
	    intro: string;
	    kind: string;
	    lastChapter: string;
	    wordCount: string;
	    bookId: string;
	    sourceKey: string;
	    source: string;
	    tab: string;
	
	    static createFrom(source: any = {}) {
	        return new Book(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.author = source["author"];
	        this.cover = source["cover"];
	        this.intro = source["intro"];
	        this.kind = source["kind"];
	        this.lastChapter = source["lastChapter"];
	        this.wordCount = source["wordCount"];
	        this.bookId = source["bookId"];
	        this.sourceKey = source["sourceKey"];
	        this.source = source["source"];
	        this.tab = source["tab"];
	    }
	}
	export class Chapter {
	    title: string;
	    itemId: string;
	
	    static createFrom(source: any = {}) {
	        return new Chapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.itemId = source["itemId"];
	    }
	}
	export class ChapterContent {
	    title: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ChapterContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.content = source["content"];
	    }
	}

}

