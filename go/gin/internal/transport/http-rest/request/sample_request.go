package request

/**
 * To check whether it is an empty value or not after unmarshal, a pointer type is used.
 * If it is not a pointer type, the default value of the corresponding data type is entered.
 * e.g. number -> 0, string -> "", array -> ["",""], slice -> []
 *
 * cf) It is essential to check for nil before referencing the converted data later.
 */
