import json

def record_metadata(record):
    # Only structural metadata; never echo source/tool inputs, text, auth or errors.
    meta = {"type": record.get("type"), "subtype": record.get("subtype")}
    event = record.get("event")
    if isinstance(event, dict):
        meta["eventType"] = event.get("type")
        block = event.get("content_block")
        if isinstance(block, dict):
            meta["blockType"] = block.get("type")
            if block.get("type") == "tool_use": meta["toolName"] = block.get("name")
        delta = event.get("delta")
        if isinstance(delta, dict): meta["deltaType"] = delta.get("type")
    message = record.get("message")
    if isinstance(message, dict) and isinstance(message.get("content"), list):
        blocks = [b for b in message["content"] if isinstance(b, dict)]
        meta["blockTypes"] = [b.get("type") for b in blocks]
        meta["toolNames"] = [b.get("name") for b in blocks if b.get("type") == "tool_use"]
        meta["toolErrorCount"] = sum(b.get("type") == "tool_result" and b.get("is_error") is True for b in blocks)
    if record.get("type") == "result":
        meta["is_error"] = record.get("is_error")
        meta["resultTextBytes"] = len(record.get("result", "").encode()) if isinstance(record.get("result"), str) else 0
        for name in ["duration_ms", "duration_api_ms", "num_turns", "stop_reason"]:
            if name in record: meta[name] = record[name]
    return meta

def classify(results, *, exit_code, timed_out, parse_errors, source_unchanged, packet_unchanged, remaining, cleanup_required, wrapper_error):
    # A complete review is not an approval; root must disposition findings.
    if timed_out: return "timeout"
    if wrapper_error: return "wrapper_error"
    if remaining or cleanup_required: return "process_cleanup_failure"
    if not source_unchanged or not packet_unchanged: return "binding_changed"
    if parse_errors: return "invalid_stream"
    if len(results) != 1: return "missing_result" if not results else "multiple_results"
    result = results[0]
    subtype = result.get("subtype")
    if isinstance(subtype, str) and "max_turn" in subtype: return "max_turns"
    if result.get("is_error") is True or (isinstance(subtype, str) and subtype.startswith("error")): return "result_error"
    if exit_code != 0: return "cli_error"
    if subtype != "success" or result.get("is_error") is not False: return "unknown_result_schema"
    if result.get("permission_denials"): return "permission_denied"
    if not isinstance(result.get("result"), str) or not result["result"].strip(): return "empty_result"
    return "review_complete"
