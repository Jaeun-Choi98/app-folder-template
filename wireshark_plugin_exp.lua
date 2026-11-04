-- myproto.lua
local my_proto = Proto("myproto", "My Custom Protocol")

local msg_types = {
    [0x01] = "Heartbeat",
    [0x02] = "Status",
    [0x10] = "Command",
}

-- 필드 정의 (필터 이름, 표시 이름, 표시 형식, 값 매핑)
local f_type = ProtoField.uint8 ("myproto.type", "Type",   base.HEX, msg_types)
local f_len  = ProtoField.uint16("myproto.len",  "Length", base.DEC)
local f_data = ProtoField.bytes ("myproto.data", "Data")

my_proto.fields = { f_type, f_len, f_data }

local HEADER_LEN = 3

function my_proto.dissector(tvb, pinfo, tree)
    local buf_len = tvb:len()
    if buf_len < HEADER_LEN then return 0 end

    local msg_type = tvb(0, 1):uint()
    local data_len = tvb(1, 2):uint()      -- 리틀엔디언이면 :le_uint()

    pinfo.cols.protocol = "MYPROTO"
    pinfo.cols.info = string.format("%s (len=%d)",
        msg_types[msg_type] or "Unknown", data_len)

    local subtree = tree:add(my_proto, tvb(0, HEADER_LEN + data_len))
    subtree:add(f_type, tvb(0, 1))
    subtree:add(f_len,  tvb(1, 2))         -- 리틀엔디언이면 add_le
    if data_len > 0 and buf_len >= HEADER_LEN + data_len then
        subtree:add(f_data, tvb(HEADER_LEN, data_len))
    end

    return HEADER_LEN + data_len
end

DissectorTable.get("tcp.port"):add(9000, my_proto)