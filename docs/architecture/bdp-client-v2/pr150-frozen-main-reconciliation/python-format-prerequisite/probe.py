from pathlib import Path
import importlib.metadata,json,sys
from jsonschema import FormatChecker,Draft202012Validator
import rfc3339_validator
P=Path(__file__).parent;target=P.parent/'tools/python-format-venv'
assert sys.argv[1:]==['--released-child'] and Path(sys.prefix)==target
checker=FormatChecker()
assert 'date-time' in checker.checkers
positive='2026-10-17T12:00:00Z';negative='bad-date'
direct={'positive':checker.conforms(positive,'date-time'),'negative':checker.conforms(negative,'date-time')}
validator=Draft202012Validator({'type':'string','format':'date-time'},format_checker=checker)
positive_errors=list(validator.iter_errors(positive));negative_errors=list(validator.iter_errors(negative))
assert direct=={'positive':True,'negative':False}
assert not positive_errors and len(negative_errors)==1 and negative_errors[0].validator=='format'
provider=Path(rfc3339_validator.__file__).resolve();assert provider.is_relative_to(target.resolve())
result={'bareFormatChecker':True,'dateTimeRegistered':True,'direct':direct,'schema':{'positiveErrorCount':len(positive_errors),'negativeErrorCount':len(negative_errors),'negativeValidators':[e.validator for e in negative_errors]},'literalInterpreter':sys.executable,'prefix':sys.prefix,'providerPath':str(provider),'providerVersion':importlib.metadata.version('rfc3339-validator'),'jsonschemaVersion':importlib.metadata.version('jsonschema'),'sixVersion':importlib.metadata.version('six'),'scope':'tool prerequisite only; no example-script or validation-stage pass'}
(P/'probe'/'probe-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
