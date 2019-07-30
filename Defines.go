package zgo

//  Defines.go
//
//  Created by Tor Langballe on /5/11/15.
//

type Object int
type AnyObject int
type Dictionary map[string]interface{}

func DefinesIsRunningInSimulator() bool {
	return false
}

func DefinesIsIOS() bool {
	return false
}

func DefinesIsApple() bool {
	return false
}

func DefinesIsTVBox() bool {
	return false
}


/*
extension String {
    func removedLast(n:Int = 1)  String {
        return String(dropLast(n))
    }
}

extension Dictionary where Value : Comparable {
    func keysSortedByValue()  [Key] {
        return self.sorted{$0.1 > $1.1}.map{$0.0}
    }
}

extension Dictionary where Value : BinaryInteger {
    mutating func AdditionInsert(key:Key, add:Value) {
        if let v = self[key] {
            self[key] = v + add
        } else {
            self[key] = 1
        }
    }
}

extension Dictionary where Key : Comparable {
    func stringFromHttpParameters(escape:bool = true, sep:String="&")  String {
        let sorted = self.keys.sorted(by: {$0 < $1})
        let params = sorted.map { (k)String in
            if escape {
                let percentEscapedKey = ZStr.UrlEncode((k as! String)) ?? ""
                let percentEscapedValue = ZStr.UrlEncode((self[k] as! String)) ?? ""
                return "\(percentEscapedKey)=\(percentEscapedValue)"
            } else {
                return "\(k)=\(String(describing: self[k]))"
            }
        }
        return params.joined(separator:sep)
    }
}

// let sortedKeys = Array(dictionary.keys).sort(<) // ["A", "D", "Z"]

extension Dictionary {
    mutating func removeIf(_ check:(_ key:Key) bool) {
        if self.count > 0 {
            for (k, _) in self {
                if check(k) {
                    self.removeValue(forKey: k)
                    break
                }
            }
        }
    }
}

typealias ZRange = NSRange

extension ZRange {
    var End: Int {
        return location + length
    }
    func Contains(_ pos:Int)  bool {
        return NSLocationInRange(pos, self)
    }
}

extension Array {

    func indexWhere(_ w:(Element)bool)  Int? {
        return index(where:w)
    }

    @discardableResult mutating func removeIf(_ check:(_ object:Element) bool)  Int {
        let c = count
        self = filter { return !check($0) }
        return c - count
    }

    mutating func shuffle () {
        for i in (0 ..< self.count).reversed() {
            let ix1 = i
            let ix2 = Int(arc4random_uniform(UInt32(i+1)))
            (self[ix1], self[ix2]) = (self[ix2], self[ix1])
        }
    }

    mutating func removeRandomElement() {
        let i = Int(arc4random_uniform(UInt32(count)))
        remove(at:i)
    }

    func shuffled ()  [Element] {
        var list = Array(self)
        list.shuffle()
        return list
    }

    mutating func moveItem(from:Int, to:Int) {
        var t = to
        if from < to {
            t = Swift.max(0, t - 1) // anoth max here in Array, so use Swift.max
        }
        insert(remove(at:from), at:t)
    }
}

extension Array where Element : Equatable {
    @discardableResult mutating func addUnique(_ element:Element, atIndex:Int = -1)  bool { // adds if doesn't contain already.  returns true if adds
        for e in self {
            if e == element {
                return false
            }
        }
        if atIndex == -1 {
            append(element)
        } else {
            insert(element, at:atIndex)
        }
        return true

    }

    func head(count:Int)  Array<Element> {
        return Array(self.prefix(count))
    }

    @discardableResult mutating func appendUnique(_ elements:[Element])  Int { // adds if doesn't contain already.  returns how many added
        var count = 0
        for e in elements {
            if addUnique(e) {
                count += 1
            }
        }
        return count
    }

    func subtract(_ sub:Array<Element>)  Array<Element> {
        return self.filter{!sub.contains($0)}
    }

    @discardableResult mutating func removeByValue(_ v:Element)  bool {
        if let i = indexWhere({$0 == v}) {
            remove(at:i)
            return true
        }
        return false
    }

}

extension Array {
    mutating func sortWithCondition(_ sortFunc:@escaping (_ a:Element, _ b:Element)  bool) {
        sort(by:{ a, b in sortFunc(a, b) })
    }

    func Max<T:Comparable>(_ get:(_ e:Element)T)  Element? {
        if isEmpty {
            return nil
        }
        return reduce(first!) { (r, e) in
            if get(r) < get(e) {
                return e
            }
            return r
        }
    }
    func Min<T:Comparable>(_ get:(_ e:Element)T)  Element? {
        if isEmpty {
            return nil
        }
        return reduce(first!) { (r, e) in
            if get(r) > get(e) {
                return e
            }
            return r
        }
    }
    mutating func popFirst()  Element? {
        if count == 0 {
            return nil
        }
        let e = first!
        remove(at:0)
        return e
    }
}

extension String {
    var fullNSRange: NSRange {
        get { return NSRange(location:0, length:NSString(string:self).length) }
    }

    var fullRange: Range<Index> {
        let s = self.startIndex
        let e = self.endIndex
        let range = s ..< e
        return range
    }
}

extension Character
{
    func unicodeScalarCodePoint()  UInt32
    {
        let characterString = String(self)
        let scalars = characterString.unicodeScalars

        return scalars[scalars.startIndex].value
    }
}

func += <KeyType, ValueType> (left:inout Dictionary<KeyType, ValueType>, right:Dictionary<KeyType, ValueType>) {
    for (k, v) in right {
        left.updateValue(v, forKey:k)
    }
}
private let trueNumber = NSNumber(value: true)
private let falseNumber = NSNumber(value: false)
private let trueObjCType = String(cString: trueNumber.objCType)
private let falseObjCType = String(cString: falseNumber.objCType)

func ZIsAnyObjectbool(_ a:ZAnyObject)  bool {
    if a is NSNumber {
        let objCType = String(cString: a.objCType)
        if (a.compare(trueNumber) == ComparisonResult.orderedSame && objCType == trueObjCType)
            || (a.compare(falseNumber) == ComparisonResult.orderedSame && objCType == falseObjCType){
            return true
        }
    }
    return false
}

func ZIsAnyObjectReal(_ a:ZAnyObject)  bool {
    if let f = a as? Float64 {
        if let i = a as? Int64 {
            return Int64(f) == i
        }
    }
    return false
}

*/
