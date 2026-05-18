import 'package:flutter/material.dart';
import '../models/dashboard_models.dart';
import '../services/settings_service.dart';

class PetDetailsScreen extends StatefulWidget {
  const PetDetailsScreen({
    super.key,
    required this.settingsService,
    this.initialPet,
  });

  final SettingsService settingsService;
  final PetProfile? initialPet;

  @override
  State<PetDetailsScreen> createState() => _PetDetailsScreenState();
}

class _PetDetailsScreenState extends State<PetDetailsScreen> {
  final _formKey = GlobalKey<FormState>();

  String _petName = '';
  String _species = 'Dog';
  String _breed = '';
  String _gender = 'Male';
  int _age = 0;
  double _weightKg = 0;
  double _targetPortion = 250.0;

  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    final pet = widget.initialPet;
    if (pet != null) {
      _petName = pet.name;
      _species = pet.species;
      _breed = pet.breed;
      _age = pet.ageYears;
      _weightKg = pet.weightKg;
      _targetPortion = pet.dailyFeedTargetGrams;
    }
  }

  Future<void> _saveChanges() async {
    if (!_formKey.currentState!.validate()) return;

    _formKey.currentState!.save();

    setState(() => _isLoading = true);

    try {
      await widget.settingsService.updatePetDetails({
        'name': _petName,
        'species': _species,
        'breed': _breed,
        'gender': _gender,
        'age_years': _age,
        'weight_kg': _weightKg,
        'daily_feed_target_grams': _targetPortion,
      });

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Pet details saved successfully!'),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to save: ${e.toString()}'),
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
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        iconTheme: IconThemeData(color: Colors.blue[800]),
        title: Text(
          'Pet Details',
          style: TextStyle(
            color: Colors.blue[800],
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Avatar
              Center(
                child: Stack(
                  children: [
                    CircleAvatar(
                      radius: 50,
                      backgroundColor: Colors.blue[50],
                      child: Icon(
                        Icons.pets,
                        size: 50,
                        color: Colors.blue[300],
                      ),
                    ),
                    Positioned(
                      bottom: 0,
                      right: 0,
                      child: Container(
                        padding: const EdgeInsets.all(4),
                        decoration: BoxDecoration(
                          color: Colors.blue[600],
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white, width: 2),
                        ),
                        child: const Icon(
                          Icons.camera_alt,
                          color: Colors.white,
                          size: 20,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 32),

              // Inputs
              TextFormField(
                initialValue: _petName,
                decoration: InputDecoration(
                  labelText: 'Pet Name',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                validator: (val) =>
                    val == null || val.isEmpty ? 'Required' : null,
                onSaved: (val) => _petName = val ?? '',
              ),
              const SizedBox(height: 16),

              TextFormField(
                initialValue: _species,
                decoration: InputDecoration(
                  labelText: 'Species',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                validator: (val) =>
                    val == null || val.isEmpty ? 'Required' : null,
                onSaved: (val) => _species = val ?? 'Dog',
              ),
              const SizedBox(height: 16),

              TextFormField(
                initialValue: _breed,
                decoration: InputDecoration(
                  labelText: 'Breed',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                onSaved: (val) => _breed = val ?? '',
              ),
              const SizedBox(height: 16),

              DropdownButtonFormField<String>(
                initialValue: _gender,
                decoration: InputDecoration(
                  labelText: 'Gender',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                items: ['Male', 'Female'].map((String value) {
                  return DropdownMenuItem<String>(
                    value: value,
                    child: Text(value),
                  );
                }).toList(),
                onChanged: (val) {
                  if (val != null) setState(() => _gender = val);
                },
              ),
              const SizedBox(height: 16),

              TextFormField(
                initialValue: _age > 0 ? _age.toString() : '',
                decoration: InputDecoration(
                  labelText: 'Age',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                keyboardType: TextInputType.number,
                validator: (val) {
                  if (val == null || val.isEmpty) return 'Required';
                  if (int.tryParse(val) == null) return 'Must be a number';
                  return null;
                },
                onSaved: (val) => _age = int.parse(val ?? '0'),
              ),
              const SizedBox(height: 16),

              TextFormField(
                initialValue: _weightKg > 0 ? _weightKg.toStringAsFixed(1) : '',
                decoration: InputDecoration(
                  labelText: 'Weight',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                  suffixText: 'kg',
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                validator: (val) {
                  if (val == null || val.isEmpty) return 'Required';
                  if (double.tryParse(val) == null) return 'Must be a number';
                  return null;
                },
                onSaved: (val) => _weightKg = double.parse(val ?? '0'),
              ),
              const SizedBox(height: 16),

              TextFormField(
                initialValue: _targetPortion.toStringAsFixed(0),
                decoration: InputDecoration(
                  labelText: 'Target Daily Portion',
                  helperText: 'Recommended: 250g',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                  suffixText: 'g',
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                validator: (val) {
                  if (val == null || val.isEmpty) return 'Required';
                  if (double.tryParse(val) == null) return 'Must be a number';
                  return null;
                },
                onSaved: (val) => _targetPortion = double.parse(val ?? '0'),
              ),

              const SizedBox(height: 32),

              // Submit
              SizedBox(
                height: 56,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _saveChanges,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue[600],
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 0,
                  ),
                  child: _isLoading
                      ? const SizedBox(
                          height: 24,
                          width: 24,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Text(
                          'Save Changes',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
