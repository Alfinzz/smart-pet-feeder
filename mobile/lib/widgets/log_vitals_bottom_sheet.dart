import 'package:flutter/material.dart';
import '../services/settings_service.dart';

class LogVitalsBottomSheet extends StatefulWidget {
  const LogVitalsBottomSheet({super.key, required this.settingsService});

  final SettingsService settingsService;

  static void show(BuildContext context, SettingsService service) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => LogVitalsBottomSheet(settingsService: service),
    );
  }

  @override
  State<LogVitalsBottomSheet> createState() => _LogVitalsBottomSheetState();
}

class _LogVitalsBottomSheetState extends State<LogVitalsBottomSheet> {
  final _formKey = GlobalKey<FormState>();
  
  double _weight = 0.0;
  int _activity = 0;
  double _sleep = 0.0;
  
  bool _isLoading = false;

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    
    _formKey.currentState!.save();
    setState(() => _isLoading = true);
    
    try {
      await widget.settingsService.submitVitals(_weight, _activity, _sleep);
      
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Vitals logged successfully!'),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context); // Close bottom sheet
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to log vitals: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // To handle keyboard overlap
    final bottomInset = MediaQuery.of(context).viewInsets.bottom;
    
    return Container(
      margin: const EdgeInsets.all(16.0),
      padding: EdgeInsets.only(
        top: 16.0,
        left: 24.0,
        right: 24.0,
        bottom: 24.0 + bottomInset,
      ),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.all(Radius.circular(24.0)),
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Handle
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[300],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 24),
            
            // Header
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Log Vitals',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.blue[900],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                  color: Colors.grey[500],
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Divider(height: 1),
            const SizedBox(height: 24),
            
            // Weight Input
            const Text('Weight', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            TextFormField(
              decoration: InputDecoration(
                prefixIcon: Icon(Icons.monitor_weight_outlined, color: Colors.blue[600]),
                suffixText: 'kg',
                hintText: '0.0',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                filled: true,
                fillColor: Colors.blue[50]?.withOpacity(0.5),
              ),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              validator: (val) {
                if (val == null || val.isEmpty) return 'Required';
                if (double.tryParse(val) == null) return 'Invalid number';
                return null;
              },
              onSaved: (val) => _weight = double.parse(val ?? '0'),
            ),
            const SizedBox(height: 16),
            
            // Activity Input
            const Text('Activity', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            TextFormField(
              decoration: InputDecoration(
                prefixIcon: Icon(Icons.directions_run, color: Colors.orange[600]),
                suffixText: 'mins',
                hintText: '0',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                filled: true,
                fillColor: Colors.blue[50]?.withOpacity(0.5),
              ),
              keyboardType: TextInputType.number,
              validator: (val) {
                if (val == null || val.isEmpty) return 'Required';
                if (int.tryParse(val) == null) return 'Invalid integer';
                return null;
              },
              onSaved: (val) => _activity = int.parse(val ?? '0'),
            ),
            const SizedBox(height: 16),
            
            // Sleep Input
            const Text('Sleep', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            TextFormField(
              decoration: InputDecoration(
                prefixIcon: Icon(Icons.nights_stay_outlined, color: Colors.indigo[600]),
                suffixText: 'hrs',
                hintText: '0.0',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                filled: true,
                fillColor: Colors.blue[50]?.withOpacity(0.5),
              ),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              validator: (val) {
                if (val == null || val.isEmpty) return 'Required';
                if (double.tryParse(val) == null) return 'Invalid number';
                return null;
              },
              onSaved: (val) => _sleep = double.parse(val ?? '0'),
            ),
            const SizedBox(height: 32),
            
            // Submit Button
            SizedBox(
              height: 50,
              child: ElevatedButton.icon(
                onPressed: _isLoading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.blue[700],
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  elevation: 0,
                ),
                icon: _isLoading 
                    ? const SizedBox() 
                    : const Icon(Icons.check_circle_outline, size: 20),
                label: _isLoading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                      )
                    : const Text(
                        'Submit Vitals',
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
