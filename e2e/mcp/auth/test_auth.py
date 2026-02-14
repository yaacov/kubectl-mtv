"""Auth -- verify that bearer-token authentication is enforced."""

import pytest

from conftest import (
    TEST_NAMESPACE,
    MCPToolError,
    _make_mcp_session,
    call_tool,
)

# A clearly fake token that no Kubernetes cluster would accept.
BAD_TOKEN = "bad-token-this-should-never-authenticate"

# A simple read command used to probe authentication.
GET_PROVIDER_CMD = {
    "command": "get provider",
    "flags": {"namespace": TEST_NAMESPACE},
}


@pytest.mark.order(4)
async def test_bad_token_rejected(mcp_server_process):
    """Session with an invalid bearer token must fail tool calls."""

    bad_headers = {"Authorization": f"Bearer {BAD_TOKEN}"}

    async with _make_mcp_session(headers=bad_headers) as session:
        with pytest.raises(MCPToolError):
            await call_tool(session, "mtv_read", GET_PROVIDER_CMD, verbose=0)
