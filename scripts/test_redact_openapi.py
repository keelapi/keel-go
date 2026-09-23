"""Regression check: redaction must not remove structural schema properties."""

import unittest

from redact_openapi import redact


class RedactOpenAPITest(unittest.TestCase):
    def test_property_names_that_match_prose_keywords_remain(self) -> None:
        schema = {
            "description": "private prose",
            "properties": {
                "description": {"type": "string", "description": "field prose"},
                "tags": {"type": "array", "items": {"type": "string"}},
                "examples": {"type": "object"},
            },
        }
        redact(schema)
        self.assertEqual(set(schema["properties"]), {"description", "tags", "examples"})
        self.assertEqual(schema["properties"]["description"], {"type": "string"})
        self.assertNotIn("description", schema)


if __name__ == "__main__":
    unittest.main()
