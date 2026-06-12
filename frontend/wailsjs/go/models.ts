export namespace clientpb {
	
	export class Loot {
	    ID?: string;
	    Name?: string;
	    FileType?: number;
	    OriginHostUUID?: string;
	    Size?: number;
	    File?: commonpb.File;
	
	    static createFrom(source: any = {}) {
	        return new Loot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.FileType = source["FileType"];
	        this.OriginHostUUID = source["OriginHostUUID"];
	        this.Size = source["Size"];
	        this.File = this.convertValues(source["File"], commonpb.File);
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
	export class AllLoot {
	    Loot?: Loot[];
	
	    static createFrom(source: any = {}) {
	        return new AllLoot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Loot = this.convertValues(source["Loot"], Loot);
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
	export class Beacon {
	    ID?: string;
	    Name?: string;
	    Hostname?: string;
	    UUID?: string;
	    Username?: string;
	    UID?: string;
	    GID?: string;
	    OS?: string;
	    Arch?: string;
	    Transport?: string;
	    RemoteAddress?: string;
	    PID?: number;
	    Filename?: string;
	    LastCheckin?: number;
	    ActiveC2?: string;
	    Version?: string;
	    Evasion?: boolean;
	    IsDead?: boolean;
	    ProxyURL?: string;
	    ReconnectInterval?: number;
	    Interval?: number;
	    Jitter?: number;
	    Burned?: boolean;
	    NextCheckin?: number;
	    TasksCount?: number;
	    TasksCountCompleted?: number;
	    Locale?: string;
	    FirstContact?: number;
	    Integrity?: string;
	
	    static createFrom(source: any = {}) {
	        return new Beacon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Hostname = source["Hostname"];
	        this.UUID = source["UUID"];
	        this.Username = source["Username"];
	        this.UID = source["UID"];
	        this.GID = source["GID"];
	        this.OS = source["OS"];
	        this.Arch = source["Arch"];
	        this.Transport = source["Transport"];
	        this.RemoteAddress = source["RemoteAddress"];
	        this.PID = source["PID"];
	        this.Filename = source["Filename"];
	        this.LastCheckin = source["LastCheckin"];
	        this.ActiveC2 = source["ActiveC2"];
	        this.Version = source["Version"];
	        this.Evasion = source["Evasion"];
	        this.IsDead = source["IsDead"];
	        this.ProxyURL = source["ProxyURL"];
	        this.ReconnectInterval = source["ReconnectInterval"];
	        this.Interval = source["Interval"];
	        this.Jitter = source["Jitter"];
	        this.Burned = source["Burned"];
	        this.NextCheckin = source["NextCheckin"];
	        this.TasksCount = source["TasksCount"];
	        this.TasksCountCompleted = source["TasksCountCompleted"];
	        this.Locale = source["Locale"];
	        this.FirstContact = source["FirstContact"];
	        this.Integrity = source["Integrity"];
	    }
	}
	export class BeaconTask {
	    ID?: string;
	    BeaconID?: string;
	    CreatedAt?: number;
	    State?: string;
	    SentAt?: number;
	    CompletedAt?: number;
	    Request?: number[];
	    Response?: number[];
	    Description?: string;
	
	    static createFrom(source: any = {}) {
	        return new BeaconTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.BeaconID = source["BeaconID"];
	        this.CreatedAt = source["CreatedAt"];
	        this.State = source["State"];
	        this.SentAt = source["SentAt"];
	        this.CompletedAt = source["CompletedAt"];
	        this.Request = source["Request"];
	        this.Response = source["Response"];
	        this.Description = source["Description"];
	    }
	}
	export class BeaconTasks {
	    BeaconID?: string;
	    Tasks?: BeaconTask[];
	
	    static createFrom(source: any = {}) {
	        return new BeaconTasks(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BeaconID = source["BeaconID"];
	        this.Tasks = this.convertValues(source["Tasks"], BeaconTask);
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
	export class Beacons {
	    Beacons?: Beacon[];
	
	    static createFrom(source: any = {}) {
	        return new Beacons(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Beacons = this.convertValues(source["Beacons"], Beacon);
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
	export class CertificateData {
	    CN?: string;
	    CreationTime?: string;
	    ValidityStart?: string;
	    ValidityExpiry?: string;
	    Type?: string;
	    KeyAlgorithm?: string;
	    ID?: string;
	
	    static createFrom(source: any = {}) {
	        return new CertificateData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CN = source["CN"];
	        this.CreationTime = source["CreationTime"];
	        this.ValidityStart = source["ValidityStart"];
	        this.ValidityExpiry = source["ValidityExpiry"];
	        this.Type = source["Type"];
	        this.KeyAlgorithm = source["KeyAlgorithm"];
	        this.ID = source["ID"];
	    }
	}
	export class CertificateInfo {
	    info?: CertificateData[];
	
	    static createFrom(source: any = {}) {
	        return new CertificateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.info = this.convertValues(source["info"], CertificateData);
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
	export class Credential {
	    ID?: string;
	    Username?: string;
	    Plaintext?: string;
	    Hash?: string;
	    HashType?: number;
	    IsCracked?: boolean;
	    OriginHostUUID?: string;
	    Collection?: string;
	
	    static createFrom(source: any = {}) {
	        return new Credential(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Username = source["Username"];
	        this.Plaintext = source["Plaintext"];
	        this.Hash = source["Hash"];
	        this.HashType = source["HashType"];
	        this.IsCracked = source["IsCracked"];
	        this.OriginHostUUID = source["OriginHostUUID"];
	        this.Collection = source["Collection"];
	    }
	}
	export class Credentials {
	    Credentials?: Credential[];
	
	    static createFrom(source: any = {}) {
	        return new Credentials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Credentials = this.convertValues(source["Credentials"], Credential);
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
	export class ImplantBuild {
	    ID?: string;
	    Name?: string;
	    MD5?: string;
	    SHA1?: string;
	    SHA256?: string;
	    Burned?: boolean;
	    ImplantID?: number;
	    ImplantConfigID?: string;
	    AgeServerPublicKey?: string;
	    PeerPublicKey?: string;
	    PeerPrivateKey?: string;
	    PeerPublicKeySignature?: string;
	    MinisignServerPublicKey?: string;
	    PeerPublicKeyDigest?: string;
	    WGImplantPrivKey?: string;
	    WGServerPubKey?: string;
	    MtlsCACert?: string;
	    MtlsCert?: string;
	    MtlsKey?: string;
	    Stage?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImplantBuild(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.MD5 = source["MD5"];
	        this.SHA1 = source["SHA1"];
	        this.SHA256 = source["SHA256"];
	        this.Burned = source["Burned"];
	        this.ImplantID = source["ImplantID"];
	        this.ImplantConfigID = source["ImplantConfigID"];
	        this.AgeServerPublicKey = source["AgeServerPublicKey"];
	        this.PeerPublicKey = source["PeerPublicKey"];
	        this.PeerPrivateKey = source["PeerPrivateKey"];
	        this.PeerPublicKeySignature = source["PeerPublicKeySignature"];
	        this.MinisignServerPublicKey = source["MinisignServerPublicKey"];
	        this.PeerPublicKeyDigest = source["PeerPublicKeyDigest"];
	        this.WGImplantPrivKey = source["WGImplantPrivKey"];
	        this.WGServerPubKey = source["WGServerPubKey"];
	        this.MtlsCACert = source["MtlsCACert"];
	        this.MtlsCert = source["MtlsCert"];
	        this.MtlsKey = source["MtlsKey"];
	        this.Stage = source["Stage"];
	    }
	}
	export class ResourceID {
	    ID?: string;
	    Type?: string;
	    Name?: string;
	    Value?: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceID(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Type = source["Type"];
	        this.Name = source["Name"];
	        this.Value = source["Value"];
	    }
	}
	export class ShellcodeConfig {
	    Entropy?: number;
	    Compress?: number;
	    ExitOpt?: number;
	    Bypass?: number;
	    Headers?: number;
	    Thread?: boolean;
	    Unicode?: boolean;
	    OEP?: number;
	
	    static createFrom(source: any = {}) {
	        return new ShellcodeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Entropy = source["Entropy"];
	        this.Compress = source["Compress"];
	        this.ExitOpt = source["ExitOpt"];
	        this.Bypass = source["Bypass"];
	        this.Headers = source["Headers"];
	        this.Thread = source["Thread"];
	        this.Unicode = source["Unicode"];
	        this.OEP = source["OEP"];
	    }
	}
	export class ImplantC2 {
	    ID?: string;
	    Priority?: number;
	    URL?: string;
	    Options?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImplantC2(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Priority = source["Priority"];
	        this.URL = source["URL"];
	        this.Options = source["Options"];
	    }
	}
	export class ImplantConfig {
	    ID?: string;
	    ImplantBuilds?: ImplantBuild[];
	    ImplantProfileID?: string;
	    IsBeacon?: boolean;
	    BeaconInterval?: number;
	    BeaconJitter?: number;
	    GOOS?: string;
	    GOARCH?: string;
	    Debug?: boolean;
	    Evasion?: boolean;
	    ObfuscateSymbols?: boolean;
	    TemplateName?: string;
	    SGNEnabled?: boolean;
	    GoPackage?: string;
	    IncludeMTLS?: boolean;
	    IncludeHTTP?: boolean;
	    IncludeWG?: boolean;
	    IncludeDNS?: boolean;
	    IncludeNamePipe?: boolean;
	    IncludeTCP?: boolean;
	    WGPeerTunIP?: string;
	    WGKeyExchangePort?: number;
	    WGTcpCommsPort?: number;
	    ReconnectInterval?: number;
	    MaxConnectionErrors?: number;
	    PollTimeout?: number;
	    C2?: ImplantC2[];
	    CanaryDomains?: string[];
	    ConnectionStrategy?: string;
	    LimitDomainJoined?: boolean;
	    LimitDatetime?: string;
	    LimitHostname?: string;
	    LimitUsername?: string;
	    LimitFileExists?: string;
	    LimitLocale?: string;
	    Format?: number;
	    IsSharedLib?: boolean;
	    IsService?: boolean;
	    IsShellcode?: boolean;
	    RunAtLoad?: boolean;
	    DebugFile?: string;
	    exports?: string[];
	    ShellcodeConfig?: ShellcodeConfig;
	    ShellcodeEncoder?: number;
	    HTTPC2ConfigName?: string;
	    NetGoEnabled?: boolean;
	    TrafficEncodersEnabled?: boolean;
	    TrafficEncoders?: string[];
	    Extension?: string;
	    Assets?: commonpb.File[];
	
	    static createFrom(source: any = {}) {
	        return new ImplantConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.ImplantBuilds = this.convertValues(source["ImplantBuilds"], ImplantBuild);
	        this.ImplantProfileID = source["ImplantProfileID"];
	        this.IsBeacon = source["IsBeacon"];
	        this.BeaconInterval = source["BeaconInterval"];
	        this.BeaconJitter = source["BeaconJitter"];
	        this.GOOS = source["GOOS"];
	        this.GOARCH = source["GOARCH"];
	        this.Debug = source["Debug"];
	        this.Evasion = source["Evasion"];
	        this.ObfuscateSymbols = source["ObfuscateSymbols"];
	        this.TemplateName = source["TemplateName"];
	        this.SGNEnabled = source["SGNEnabled"];
	        this.GoPackage = source["GoPackage"];
	        this.IncludeMTLS = source["IncludeMTLS"];
	        this.IncludeHTTP = source["IncludeHTTP"];
	        this.IncludeWG = source["IncludeWG"];
	        this.IncludeDNS = source["IncludeDNS"];
	        this.IncludeNamePipe = source["IncludeNamePipe"];
	        this.IncludeTCP = source["IncludeTCP"];
	        this.WGPeerTunIP = source["WGPeerTunIP"];
	        this.WGKeyExchangePort = source["WGKeyExchangePort"];
	        this.WGTcpCommsPort = source["WGTcpCommsPort"];
	        this.ReconnectInterval = source["ReconnectInterval"];
	        this.MaxConnectionErrors = source["MaxConnectionErrors"];
	        this.PollTimeout = source["PollTimeout"];
	        this.C2 = this.convertValues(source["C2"], ImplantC2);
	        this.CanaryDomains = source["CanaryDomains"];
	        this.ConnectionStrategy = source["ConnectionStrategy"];
	        this.LimitDomainJoined = source["LimitDomainJoined"];
	        this.LimitDatetime = source["LimitDatetime"];
	        this.LimitHostname = source["LimitHostname"];
	        this.LimitUsername = source["LimitUsername"];
	        this.LimitFileExists = source["LimitFileExists"];
	        this.LimitLocale = source["LimitLocale"];
	        this.Format = source["Format"];
	        this.IsSharedLib = source["IsSharedLib"];
	        this.IsService = source["IsService"];
	        this.IsShellcode = source["IsShellcode"];
	        this.RunAtLoad = source["RunAtLoad"];
	        this.DebugFile = source["DebugFile"];
	        this.exports = source["exports"];
	        this.ShellcodeConfig = this.convertValues(source["ShellcodeConfig"], ShellcodeConfig);
	        this.ShellcodeEncoder = source["ShellcodeEncoder"];
	        this.HTTPC2ConfigName = source["HTTPC2ConfigName"];
	        this.NetGoEnabled = source["NetGoEnabled"];
	        this.TrafficEncodersEnabled = source["TrafficEncodersEnabled"];
	        this.TrafficEncoders = source["TrafficEncoders"];
	        this.Extension = source["Extension"];
	        this.Assets = this.convertValues(source["Assets"], commonpb.File);
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
	export class ImplantBuilds {
	    Configs?: Record<string, ImplantConfig>;
	    ResourceIDs?: Record<string, ResourceID>;
	    staged?: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new ImplantBuilds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Configs = this.convertValues(source["Configs"], ImplantConfig, true);
	        this.ResourceIDs = this.convertValues(source["ResourceIDs"], ResourceID, true);
	        this.staged = source["staged"];
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
	
	
	export class ImplantProfile {
	    ID?: string;
	    Name?: string;
	    Config?: ImplantConfig;
	
	    static createFrom(source: any = {}) {
	        return new ImplantProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Config = this.convertValues(source["Config"], ImplantConfig);
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
	export class ImplantProfiles {
	    Profiles?: ImplantProfile[];
	
	    static createFrom(source: any = {}) {
	        return new ImplantProfiles(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Profiles = this.convertValues(source["Profiles"], ImplantProfile);
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
	export class Job {
	    ID?: number;
	    Name?: string;
	    Description?: string;
	    Protocol?: string;
	    Port?: number;
	    Domains?: string[];
	    ProfileName?: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Protocol = source["Protocol"];
	        this.Port = source["Port"];
	        this.Domains = source["Domains"];
	        this.ProfileName = source["ProfileName"];
	    }
	}
	export class Jobs {
	    Active?: Job[];
	
	    static createFrom(source: any = {}) {
	        return new Jobs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Active = this.convertValues(source["Active"], Job);
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
	
	export class Operator {
	    Online?: boolean;
	    Name?: string;
	
	    static createFrom(source: any = {}) {
	        return new Operator(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Online = source["Online"];
	        this.Name = source["Name"];
	    }
	}
	export class Operators {
	    Operators?: Operator[];
	
	    static createFrom(source: any = {}) {
	        return new Operators(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Operators = this.convertValues(source["Operators"], Operator);
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
	export class Session {
	    ID?: string;
	    Name?: string;
	    Hostname?: string;
	    UUID?: string;
	    Username?: string;
	    UID?: string;
	    GID?: string;
	    OS?: string;
	    Arch?: string;
	    Transport?: string;
	    RemoteAddress?: string;
	    PID?: number;
	    Filename?: string;
	    LastCheckin?: number;
	    ActiveC2?: string;
	    Version?: string;
	    Evasion?: boolean;
	    IsDead?: boolean;
	    ReconnectInterval?: number;
	    ProxyURL?: string;
	    Burned?: boolean;
	    Extensions?: string[];
	    PeerID?: number;
	    Locale?: string;
	    FirstContact?: number;
	    Integrity?: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Hostname = source["Hostname"];
	        this.UUID = source["UUID"];
	        this.Username = source["Username"];
	        this.UID = source["UID"];
	        this.GID = source["GID"];
	        this.OS = source["OS"];
	        this.Arch = source["Arch"];
	        this.Transport = source["Transport"];
	        this.RemoteAddress = source["RemoteAddress"];
	        this.PID = source["PID"];
	        this.Filename = source["Filename"];
	        this.LastCheckin = source["LastCheckin"];
	        this.ActiveC2 = source["ActiveC2"];
	        this.Version = source["Version"];
	        this.Evasion = source["Evasion"];
	        this.IsDead = source["IsDead"];
	        this.ReconnectInterval = source["ReconnectInterval"];
	        this.ProxyURL = source["ProxyURL"];
	        this.Burned = source["Burned"];
	        this.Extensions = source["Extensions"];
	        this.PeerID = source["PeerID"];
	        this.Locale = source["Locale"];
	        this.FirstContact = source["FirstContact"];
	        this.Integrity = source["Integrity"];
	    }
	}
	export class PivotGraphEntry {
	    PeerID?: number;
	    Session?: Session;
	    Name?: string;
	    Children?: PivotGraphEntry[];
	
	    static createFrom(source: any = {}) {
	        return new PivotGraphEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PeerID = source["PeerID"];
	        this.Session = this.convertValues(source["Session"], Session);
	        this.Name = source["Name"];
	        this.Children = this.convertValues(source["Children"], PivotGraphEntry);
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
	export class PivotGraph {
	    Children?: PivotGraphEntry[];
	
	    static createFrom(source: any = {}) {
	        return new PivotGraph(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Children = this.convertValues(source["Children"], PivotGraphEntry);
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
	
	
	
	export class Sessions {
	    Sessions?: Session[];
	
	    static createFrom(source: any = {}) {
	        return new Sessions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Sessions = this.convertValues(source["Sessions"], Session);
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
	
	export class Version {
	    Major?: number;
	    Minor?: number;
	    Patch?: number;
	    Commit?: string;
	    Dirty?: boolean;
	    CompiledAt?: number;
	    OS?: string;
	    Arch?: string;
	
	    static createFrom(source: any = {}) {
	        return new Version(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Major = source["Major"];
	        this.Minor = source["Minor"];
	        this.Patch = source["Patch"];
	        this.Commit = source["Commit"];
	        this.Dirty = source["Dirty"];
	        this.CompiledAt = source["CompiledAt"];
	        this.OS = source["OS"];
	        this.Arch = source["Arch"];
	    }
	}
	export class WebContent {
	    ID?: string;
	    WebsiteID?: string;
	    Path?: string;
	    ContentType?: string;
	    Size?: number;
	    OriginalFile?: string;
	    Sha256?: string;
	    Content?: number[];
	
	    static createFrom(source: any = {}) {
	        return new WebContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.WebsiteID = source["WebsiteID"];
	        this.Path = source["Path"];
	        this.ContentType = source["ContentType"];
	        this.Size = source["Size"];
	        this.OriginalFile = source["OriginalFile"];
	        this.Sha256 = source["Sha256"];
	        this.Content = source["Content"];
	    }
	}
	export class Website {
	    ID?: string;
	    Name?: string;
	    Contents?: Record<string, WebContent>;
	
	    static createFrom(source: any = {}) {
	        return new Website(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Contents = this.convertValues(source["Contents"], WebContent, true);
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
	export class Websites {
	    Websites?: Website[];
	
	    static createFrom(source: any = {}) {
	        return new Websites(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Websites = this.convertValues(source["Websites"], Website);
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

export namespace commonpb {
	
	export class File {
	    Name?: string;
	    Data?: number[];
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Data = source["Data"];
	    }
	}
	export class Process {
	    Pid?: number;
	    Ppid?: number;
	    Executable?: string;
	    Owner?: string;
	    Architecture?: string;
	    SessionID?: number;
	    CmdLine?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pid = source["Pid"];
	        this.Ppid = source["Ppid"];
	        this.Executable = source["Executable"];
	        this.Owner = source["Owner"];
	        this.Architecture = source["Architecture"];
	        this.SessionID = source["SessionID"];
	        this.CmdLine = source["CmdLine"];
	    }
	}
	export class Response {
	    Err?: string;
	    Async?: boolean;
	    BeaconID?: string;
	    TaskID?: string;
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Err = source["Err"];
	        this.Async = source["Async"];
	        this.BeaconID = source["BeaconID"];
	        this.TaskID = source["TaskID"];
	    }
	}

}

export namespace main {
	
	export class AutomationFilter {
	    os: string;
	    arch: string;
	    hostname: string;
	    username: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new AutomationFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.hostname = source["hostname"];
	        this.username = source["username"];
	        this.name = source["name"];
	    }
	}
	export class AutomationRule {
	    id: string;
	    name: string;
	    description: string;
	    enabled: boolean;
	    trigger: string;
	    targetKind: string;
	    filter: AutomationFilter;
	    executionMode: string;
	    commands: string[];
	    script: string;
	    timeoutSeconds: number;
	    continueOnError: boolean;
	    delaySeconds: number;
	    cooldownSeconds: number;
	    intervalSeconds: number;
	    maxRuns: number;
	    runCount: number;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AutomationRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.trigger = source["trigger"];
	        this.targetKind = source["targetKind"];
	        this.filter = this.convertValues(source["filter"], AutomationFilter);
	        this.executionMode = source["executionMode"];
	        this.commands = source["commands"];
	        this.script = source["script"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.continueOnError = source["continueOnError"];
	        this.delaySeconds = source["delaySeconds"];
	        this.cooldownSeconds = source["cooldownSeconds"];
	        this.intervalSeconds = source["intervalSeconds"];
	        this.maxRuns = source["maxRuns"];
	        this.runCount = source["runCount"];
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
	export class AutomationRun {
	    id: string;
	    ruleId: string;
	    ruleName: string;
	    trigger: string;
	    targetId: string;
	    targetName: string;
	    targetKind: string;
	    commands: string[];
	    output: string;
	    error: string;
	    status: string;
	    startedAt: number;
	    finishedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AutomationRun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ruleId = source["ruleId"];
	        this.ruleName = source["ruleName"];
	        this.trigger = source["trigger"];
	        this.targetId = source["targetId"];
	        this.targetName = source["targetName"];
	        this.targetKind = source["targetKind"];
	        this.commands = source["commands"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.status = source["status"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	    }
	}
	export class CommandArg {
	    name: string;
	    required: boolean;
	    variadic: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommandArg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.required = source["required"];
	        this.variadic = source["variadic"];
	    }
	}
	export class CommandFlag {
	    name: string;
	    shorthand?: string;
	    usage: string;
	    type: string;
	    default?: string;
	    required: boolean;
	    boolean: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommandFlag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.shorthand = source["shorthand"];
	        this.usage = source["usage"];
	        this.type = source["type"];
	        this.default = source["default"];
	        this.required = source["required"];
	        this.boolean = source["boolean"];
	    }
	}
	export class CommandSchema {
	    name: string;
	    path: string;
	    usage: string;
	    description: string;
	    arguments: CommandArg[];
	    flags: CommandFlag[];
	    needsInput: boolean;
	    supported: boolean;
	    unavailable?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.usage = source["usage"];
	        this.description = source["description"];
	        this.arguments = this.convertValues(source["arguments"], CommandArg);
	        this.flags = this.convertValues(source["flags"], CommandFlag);
	        this.needsInput = source["needsInput"];
	        this.supported = source["supported"];
	        this.unavailable = source["unavailable"];
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
	export class CommandGroup {
	    id: string;
	    title: string;
	    commands: CommandSchema[];
	
	    static createFrom(source: any = {}) {
	        return new CommandGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.commands = this.convertValues(source["commands"], CommandSchema);
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
	export class CommandCatalog {
	    scope: string;
	    groups: CommandGroup[];
	
	    static createFrom(source: any = {}) {
	        return new CommandCatalog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scope = source["scope"];
	        this.groups = this.convertValues(source["groups"], CommandGroup);
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
	
	
	
	export class PivotConnectionSnapshot {
	    PeerID: number;
	    RemoteAddress: string;
	
	    static createFrom(source: any = {}) {
	        return new PivotConnectionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PeerID = source["PeerID"];
	        this.RemoteAddress = source["RemoteAddress"];
	    }
	}
	export class PivotListenerSnapshot {
	    ParentSessionID: string;
	    ID: number;
	    Type: string;
	    BindAddress: string;
	    Pivots: PivotConnectionSnapshot[];
	
	    static createFrom(source: any = {}) {
	        return new PivotListenerSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ParentSessionID = source["ParentSessionID"];
	        this.ID = source["ID"];
	        this.Type = source["Type"];
	        this.BindAddress = source["BindAddress"];
	        this.Pivots = this.convertValues(source["Pivots"], PivotConnectionSnapshot);
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
	export class RegistryValue {
	    name: string;
	    type: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new RegistryValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.value = source["value"];
	    }
	}
	export class ShellInfo {
	    id: string;
	    sessionID: string;
	    path: string;
	    pid: number;
	    pty: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ShellInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionID = source["sessionID"];
	        this.path = source["path"];
	        this.pid = source["pid"];
	        this.pty = source["pty"];
	    }
	}

}

export namespace sliverpb {
	
	export class FileInfo {
	    Name?: string;
	    IsDir?: boolean;
	    Size?: number;
	    ModTime?: number;
	    Mode?: string;
	    Link?: string;
	    Uid?: string;
	    Gid?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.IsDir = source["IsDir"];
	        this.Size = source["Size"];
	        this.ModTime = source["ModTime"];
	        this.Mode = source["Mode"];
	        this.Link = source["Link"];
	        this.Uid = source["Uid"];
	        this.Gid = source["Gid"];
	    }
	}
	export class Ls {
	    Path?: string;
	    Exists?: boolean;
	    Files?: FileInfo[];
	    timezone?: string;
	    timezoneOffset?: number;
	    Response?: commonpb.Response;
	
	    static createFrom(source: any = {}) {
	        return new Ls(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Exists = source["Exists"];
	        this.Files = this.convertValues(source["Files"], FileInfo);
	        this.timezone = source["timezone"];
	        this.timezoneOffset = source["timezoneOffset"];
	        this.Response = this.convertValues(source["Response"], commonpb.Response);
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
	export class Ps {
	    Processes?: commonpb.Process[];
	    Response?: commonpb.Response;
	
	    static createFrom(source: any = {}) {
	        return new Ps(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Processes = this.convertValues(source["Processes"], commonpb.Process);
	        this.Response = this.convertValues(source["Response"], commonpb.Response);
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
	export class RegistrySubKeyList {
	    Subkeys?: string[];
	    Response?: commonpb.Response;
	
	    static createFrom(source: any = {}) {
	        return new RegistrySubKeyList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Subkeys = source["Subkeys"];
	        this.Response = this.convertValues(source["Response"], commonpb.Response);
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
	export class RegistryValuesList {
	    ValueNames?: string[];
	    Response?: commonpb.Response;
	
	    static createFrom(source: any = {}) {
	        return new RegistryValuesList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ValueNames = source["ValueNames"];
	        this.Response = this.convertValues(source["Response"], commonpb.Response);
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

