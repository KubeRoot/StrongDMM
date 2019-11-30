package strongdmm.util.extension

import gnu.trove.map.hash.TIntObjectHashMap
import gnu.trove.map.hash.TObjectIntHashMap
import gnu.trove.map.hash.TShortObjectHashMap

inline fun <T> TIntObjectHashMap<T>.getOrPut(key: Int, obj: (Int) -> T): T {
    if (!containsKey(key)) {
        put(key, obj(key))
    }
    return this[key]
}

inline fun <T> TObjectIntHashMap<T>.getOrPut(key: T, obj: (T) -> Int): Int {
    if (!containsKey(key)) {
        put(key, obj(key))
    }
    return this[key]
}

inline fun <T> TShortObjectHashMap<T>.getOrPut(key: Short, obj: (Short) -> T): T {
    if (!containsKey(key)) {
        put(key, obj(key))
    }
    return this[key]
}
